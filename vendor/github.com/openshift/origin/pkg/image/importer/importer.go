package importer

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/golang/glog"
	gocontext "golang.org/x/net/context"

	"github.com/docker/distribution"
	"github.com/docker/distribution/context"
	"github.com/docker/distribution/digest"
	"github.com/docker/distribution/manifest/schema1"
	"github.com/docker/distribution/reference"
	"github.com/docker/distribution/registry/api/errcode"
	"github.com/docker/distribution/registry/api/v2"
	registryclient "github.com/docker/distribution/registry/client"
	"github.com/docker/distribution/registry/client/auth"
	"github.com/docker/distribution/registry/client/transport"

	kapi "k8s.io/kubernetes/pkg/api"
	kapierrors "k8s.io/kubernetes/pkg/api/errors"
	"k8s.io/kubernetes/pkg/api/unversioned"
	"k8s.io/kubernetes/pkg/util"
	"k8s.io/kubernetes/pkg/util/sets"
	"k8s.io/kubernetes/pkg/util/validation/field"

	"github.com/openshift/origin/pkg/dockerregistry"
	"github.com/openshift/origin/pkg/image/api"
	"github.com/openshift/origin/pkg/image/api/dockerpre012"
)

// Add a dockerregistry.Client to the passed context with this key to support v1 Docker registry importing
const ContextKeyV1RegistryClient = "v1-registry-client"

// Interface loads images into an image stream import request.
type Interface interface {
	Import(ctx gocontext.Context, isi *api.ImageStreamImport) error
}

// RepositoryRetriever fetches a Docker distribution.Repository.
type RepositoryRetriever interface {
	// Repository returns a properly authenticated distribution.Repository for the given registry, repository
	// name, and insecure toleration behavior.
	Repository(ctx gocontext.Context, registry *url.URL, repoName string, insecure bool) (distribution.Repository, error)
}

// ErrNotV2Registry is returned when the server does not report itself as a V2 Docker registry
type ErrNotV2Registry struct {
	Registry string
}

func (e *ErrNotV2Registry) Error() string {
	return fmt.Sprintf("endpoint %q does not support v2 API", e.Registry)
}

// ImageStreamImport implements an import strategy for Docker images. It keeps a cache of images
// per distinct auth context to reduce duplicate loads. This type is not thread safe.
type ImageStreamImporter struct {
	maximumTagsPerRepo int

	retriever RepositoryRetriever
	limiter   util.RateLimiter

	digestToRepositoryCache map[gocontext.Context]map[manifestKey]*api.Image
}

// NewImageStreamImport creates an importer that will load images from a remote Docker registry into an
// ImageStreamImport object. Limiter may be nil.
func NewImageStreamImporter(retriever RepositoryRetriever, maximumTagsPerRepo int, limiter util.RateLimiter) *ImageStreamImporter {
	if limiter == nil {
		limiter = util.NewFakeAlwaysRateLimiter()
	}
	return &ImageStreamImporter{
		maximumTagsPerRepo: maximumTagsPerRepo,

		retriever: retriever,
		limiter:   limiter,

		digestToRepositoryCache: make(map[gocontext.Context]map[manifestKey]*api.Image),
	}
}

// contextImageCache returns the image cache entry for a context.
func (i *ImageStreamImporter) contextImageCache(ctx gocontext.Context) map[manifestKey]*api.Image {
	cache := i.digestToRepositoryCache[ctx]
	if cache == nil {
		cache = make(map[manifestKey]*api.Image)
		i.digestToRepositoryCache[ctx] = cache
	}
	return cache
}

// Import tries to complete the provided isi object with images loaded from remote registries.
func (i *ImageStreamImporter) Import(ctx gocontext.Context, isi *api.ImageStreamImport) error {
	cache := i.contextImageCache(ctx)
	importImages(ctx, i.retriever, isi, cache, i.limiter)
	importFromRepository(ctx, i.retriever, isi, i.maximumTagsPerRepo, cache, i.limiter)
	return nil
}

// importImages updates the passed ImageStreamImport object and sets Status for each image based on whether the import
// succeeded or failed. Cache is updated with any loaded images. Limiter is optional and controls how fast images are updated.
func importImages(ctx gocontext.Context, retriever RepositoryRetriever, isi *api.ImageStreamImport, cache map[manifestKey]*api.Image, limiter util.RateLimiter) {
	tags := make(map[manifestKey][]int)
	ids := make(map[manifestKey][]int)
	repositories := make(map[repositoryKey]*importRepository)

	isi.Status.Images = make([]api.ImageImportStatus, len(isi.Spec.Images))
	for i := range isi.Spec.Images {
		spec := &isi.Spec.Images[i]
		from := spec.From
		if from.Kind != "DockerImage" {
			continue
		}
		ref, err := api.ParseDockerImageReference(from.Name)
		if err != nil {
			isi.Status.Images[i].Status = invalidStatus("", field.Invalid(field.NewPath("from", "name"), from.Name, fmt.Sprintf("invalid name: %v", err)))
			continue
		}
		defaultRef := ref.DockerClientDefaults()
		repoName := defaultRef.RepositoryName()
		registryURL := defaultRef.RegistryURL()

		key := repositoryKey{url: *registryURL, name: repoName}
		repo, ok := repositories[key]
		if !ok {
			repo = &importRepository{
				Ref:      ref,
				Registry: &key.url,
				Name:     key.name,
				Insecure: spec.ImportPolicy.Insecure,
			}
			repositories[key] = repo
		}

		if len(defaultRef.ID) > 0 {
			id := manifestKey{repositoryKey: key}
			id.value = defaultRef.ID
			ids[id] = append(ids[id], i)
			if len(ids[id]) == 1 {
				repo.Digests = append(repo.Digests, importDigest{
					Name:  defaultRef.ID,
					Image: cache[id],
				})
			}
		} else {
			tag := manifestKey{repositoryKey: key}
			tag.value = defaultRef.Tag
			tags[tag] = append(tags[tag], i)
			if len(tags[tag]) == 1 {
				repo.Tags = append(repo.Tags, importTag{
					Name:  defaultRef.Tag,
					Image: cache[tag],
				})
			}
		}
	}

	// for each repository we found, import all tags and digests
	for key, repo := range repositories {
		importRepositoryFromDocker(ctx, retriever, repo, limiter)
		for _, tag := range repo.Tags {
			j := manifestKey{repositoryKey: key}
			j.value = tag.Name
			if tag.Image != nil {
				cache[j] = tag.Image
			}
			for _, index := range tags[j] {
				if tag.Err != nil {
					setImageImportStatus(isi, index, tag.Err)
					continue
				}
				copied := *tag.Image
				image := &isi.Status.Images[index]
				ref := repo.Ref
				ref.Tag, ref.ID = tag.Name, copied.Name
				copied.DockerImageReference = ref.MostSpecific().Exact()
				image.Tag = tag.Name
				image.Image = &copied
				image.Status.Status = unversioned.StatusSuccess
			}
		}
		for _, digest := range repo.Digests {
			j := manifestKey{repositoryKey: key}
			j.value = digest.Name
			if digest.Image != nil {
				cache[j] = digest.Image
			}
			for _, index := range ids[j] {
				if digest.Err != nil {
					setImageImportStatus(isi, index, digest.Err)
					continue
				}
				image := &isi.Status.Images[index]
				copied := *digest.Image
				ref := repo.Ref
				ref.Tag, ref.ID = "", copied.Name
				copied.DockerImageReference = ref.MostSpecific().Exact()
				image.Image = &copied
				image.Status.Status = unversioned.StatusSuccess
			}
		}
	}
}

// importFromRepository imports the repository named on the ImageStreamImport, if any, importing up to maximumTags, and reporting
// status on each image that is attempted to be imported. If the repository cannot be found or tags cannot be retrieved, the repository
// status field is set.
func importFromRepository(ctx gocontext.Context, retriever RepositoryRetriever, isi *api.ImageStreamImport, maximumTags int, cache map[manifestKey]*api.Image, limiter util.RateLimiter) {
	if isi.Spec.Repository == nil {
		return
	}
	isi.Status.Repository = &api.RepositoryImportStatus{}
	status := isi.Status.Repository

	spec := isi.Spec.Repository
	from := spec.From
	if from.Kind != "DockerImage" {
		return
	}
	ref, err := api.ParseDockerImageReference(from.Name)
	if err != nil {
		status.Status = invalidStatus("", field.Invalid(field.NewPath("from", "name"), from.Name, fmt.Sprintf("invalid name: %v", err)))
		return
	}
	defaultRef := ref.DockerClientDefaults()
	repoName := defaultRef.RepositoryName()
	registryURL := defaultRef.RegistryURL()

	key := repositoryKey{url: *registryURL, name: repoName}
	repo := &importRepository{
		Ref:         ref,
		Registry:    &key.url,
		Name:        key.name,
		Insecure:    spec.ImportPolicy.Insecure,
		MaximumTags: maximumTags,
	}
	importRepositoryFromDocker(ctx, retriever, repo, limiter)

	if repo.Err != nil {
		status.Status = imageImportStatus(repo.Err, "", "repository")
		return
	}

	additional := []string{}
	tagKey := manifestKey{repositoryKey: key}
	for _, s := range repo.AdditionalTags {
		tagKey.value = s
		if image, ok := cache[tagKey]; ok {
			repo.Tags = append(repo.Tags, importTag{
				Name:  s,
				Image: image,
			})
		} else {
			additional = append(additional, s)
		}
	}
	status.AdditionalTags = additional

	failures := 0
	status.Status.Status = unversioned.StatusSuccess
	status.Images = make([]api.ImageImportStatus, len(repo.Tags))
	for i, tag := range repo.Tags {
		if tag.Err != nil {
			failures++
			status.Images[i].Status = imageImportStatus(tag.Err, "", "repository")
			continue
		}
		status.Images[i].Status.Status = unversioned.StatusSuccess

		copied := *tag.Image
		ref.Tag, ref.ID = tag.Name, copied.Name
		copied.DockerImageReference = ref.MostSpecific().Exact()
		status.Images[i].Tag = tag.Name
		status.Images[i].Image = &copied
	}
	if failures > 0 {
		status.Status.Status = unversioned.StatusFailure
		status.Status.Reason = unversioned.StatusReason("ImportFailed")
		switch failures {
		case 1:
			status.Status.Message = "one of the images from this repository failed to import"
		default:
			status.Status.Message = fmt.Sprintf("%d of the images from this repository failed to import", failures)
		}
	}
}

func applyErrorToRepository(repository *importRepository, err error) {
	repository.Err = err
	for i := range repository.Tags {
		repository.Tags[i].Err = err
	}
	for i := range repository.Digests {
		repository.Digests[i].Err = err
	}
}

// importRepositoryFromDocker loads the tags and images requested in the passed importRepository, obeying the
// optional rate limiter.  Errors are set onto the individual tags and digest objects.
func importRepositoryFromDocker(ctx gocontext.Context, retriever RepositoryRetriever, repository *importRepository, limiter util.RateLimiter) {
	glog.V(5).Infof("importing remote Docker repository registry=%s repository=%s insecure=%t", repository.Registry, repository.Name, repository.Insecure)
	// retrieve the repository
	repo, err := retriever.Repository(ctx, repository.Registry, repository.Name, repository.Insecure)
	if err != nil {
		glog.V(5).Infof("unable to access repository %#v: %#v", repository, err)
		switch {
		case err == reference.ErrReferenceInvalidFormat:
			err = field.Invalid(field.NewPath("from", "name"), repository.Name, "the provided repository name is not valid")
		case isDockerError(err, v2.ErrorCodeNameUnknown):
			err = kapierrors.NewNotFound(api.Resource("dockerimage"), repository.Ref.Exact())
		case isDockerError(err, errcode.ErrorCodeUnauthorized):
			err = kapierrors.NewUnauthorized(fmt.Sprintf("you may not have access to the Docker image %q", repository.Ref.Exact()))
		case strings.Contains(err.Error(), "tls: oversized record received with length") && !repository.Insecure:
			err = kapierrors.NewBadRequest("this repository is HTTP only and requires the insecure flag to import")
		case strings.HasSuffix(err.Error(), "no basic auth credentials"):
			err = kapierrors.NewUnauthorized(fmt.Sprintf("you may not have access to the Docker image %q and did not have credentials to the repository", repository.Ref.Exact()))
		case strings.HasSuffix(err.Error(), "does not support v2 API"):
			importRepositoryFromDockerV1(ctx, repository, limiter)
			return
		}
		applyErrorToRepository(repository, err)
		return
	}

	// get a manifest context
	s, err := repo.Manifests(ctx)
	if err != nil {
		glog.V(5).Infof("unable to access manifests for repository %#v: %#v", repository, err)
		switch {
		case isDockerError(err, v2.ErrorCodeNameUnknown):
			err = kapierrors.NewNotFound(api.Resource("dockerimage"), repository.Ref.Exact())
		case isDockerError(err, errcode.ErrorCodeUnauthorized):
			err = kapierrors.NewUnauthorized(fmt.Sprintf("you may not have access to the Docker image %q", repository.Ref.Exact()))
		case strings.HasSuffix(err.Error(), "no basic auth credentials"):
			err = kapierrors.NewUnauthorized(fmt.Sprintf("you may not have access to the Docker image %q and did not have credentials to the repository", repository.Ref.Exact()))
		}
		applyErrorToRepository(repository, err)
		return
	}

	// if repository import is requested (MaximumTags), attempt to load the tags, sort them, and request the first N
	if count := repository.MaximumTags; count > 0 || count == -1 {
		tags, err := s.Tags()
		if err != nil {
			glog.V(5).Infof("unable to access tags for repository %#v: %#v", repository, err)
			switch {
			case isDockerError(err, v2.ErrorCodeNameUnknown):
				err = kapierrors.NewNotFound(api.Resource("dockerimage"), repository.Ref.Exact())
			case isDockerError(err, errcode.ErrorCodeUnauthorized):
				err = kapierrors.NewUnauthorized(fmt.Sprintf("you may not have access to the Docker image %q", repository.Ref.Exact()))
			}
			repository.Err = err
			return
		}
		// some images on the Hub have empty tags - treat those as "latest"
		set := sets.NewString(tags...)
		if set.Has("") {
			set.Delete("")
			set.Insert(api.DefaultImageTag)
		}
		tags = set.List()
		// include only the top N tags in the result, put the rest in AdditionalTags
		api.PrioritizeTags(tags)
		for _, s := range tags {
			if count <= 0 && repository.MaximumTags != -1 {
				repository.AdditionalTags = append(repository.AdditionalTags, s)
				continue
			}
			count--
			repository.Tags = append(repository.Tags, importTag{
				Name: s,
			})
		}
	}

	// load digests
	for i := range repository.Digests {
		importDigest := &repository.Digests[i]
		if importDigest.Err != nil || importDigest.Image != nil {
			continue
		}
		d, err := digest.ParseDigest(importDigest.Name)
		if err != nil {
			importDigest.Err = err
			continue
		}
		limiter.Accept()
		m, err := s.Get(d)
		if err != nil {
			glog.V(5).Infof("unable to access digest %q for repository %#v: %#v", d, repository, err)
			switch {
			case isDockerError(err, v2.ErrorCodeManifestUnknown):
				ref := repository.Ref
				ref.Tag, ref.ID = "", importDigest.Name
				err = kapierrors.NewNotFound(api.Resource("dockerimage"), ref.Exact())
			case isDockerError(err, errcode.ErrorCodeUnauthorized):
				err = kapierrors.NewUnauthorized(fmt.Sprintf("you may not have access to the Docker image %q", repository.Ref.Exact()))
			case strings.HasSuffix(err.Error(), "no basic auth credentials"):
				err = kapierrors.NewUnauthorized(fmt.Sprintf("you may not have access to the Docker image %q", repository.Ref.Exact()))
			}
			importDigest.Err = err
			continue
		}
		importDigest.Image, err = schema1ToImage(m, d)
		if err != nil {
			importDigest.Err = err
			continue
		}
		if err := api.ImageWithMetadata(importDigest.Image); err != nil {
			importDigest.Err = err
			continue
		}
	}

	for i := range repository.Tags {
		importTag := &repository.Tags[i]
		if importTag.Err != nil || importTag.Image != nil {
			continue
		}
		limiter.Accept()
		m, err := s.GetByTag(importTag.Name)
		if err != nil {
			glog.V(5).Infof("unable to access tag %q for repository %#v: %#v", importTag.Name, repository, err)
			switch {
			case isDockerError(err, v2.ErrorCodeManifestUnknown):
				ref := repository.Ref
				ref.Tag = importTag.Name
				err = kapierrors.NewNotFound(api.Resource("dockerimage"), ref.Exact())
			case isDockerError(err, errcode.ErrorCodeUnauthorized):
				err = kapierrors.NewUnauthorized(fmt.Sprintf("you may not have access to the Docker image %q", repository.Ref.Exact()))
			case strings.HasSuffix(err.Error(), "no basic auth credentials"):
				err = kapierrors.NewUnauthorized(fmt.Sprintf("you may not have access to the Docker image %q", repository.Ref.Exact()))
			}
			importTag.Err = err
			continue
		}
		importTag.Image, err = schema1ToImage(m, "")
		if err != nil {
			importTag.Err = err
			continue
		}
		if err := api.ImageWithMetadata(importTag.Image); err != nil {
			importTag.Err = err
			continue
		}
	}
}

func importRepositoryFromDockerV1(ctx gocontext.Context, repository *importRepository, limiter util.RateLimiter) {
	value := ctx.Value(ContextKeyV1RegistryClient)
	if value == nil {
		err := kapierrors.NewForbidden(api.Resource(""), "", fmt.Errorf("registry %q does not support the v2 Registry API", repository.Registry.Host)).(*kapierrors.StatusError)
		err.ErrStatus.Reason = "NotV2Registry"
		applyErrorToRepository(repository, err)
		return
	}
	client, ok := value.(dockerregistry.Client)
	if !ok {
		err := kapierrors.NewForbidden(api.Resource(""), "", fmt.Errorf("registry %q does not support the v2 Registry API", repository.Registry.Host)).(*kapierrors.StatusError)
		err.ErrStatus.Reason = "NotV2Registry"
		return
	}
	conn, err := client.Connect(repository.Registry.Host, repository.Insecure)
	if err != nil {
		applyErrorToRepository(repository, err)
		return
	}

	// if repository import is requested (MaximumTags), attempt to load the tags, sort them, and request the first N
	if count := repository.MaximumTags; count > 0 {
		tagMap, err := conn.ImageTags(repository.Ref.Namespace, repository.Ref.Name)
		if err != nil {
			repository.Err = err
			return
		}
		tags := make([]string, 0, len(tagMap))
		for tag := range tagMap {
			tags = append(tags, tag)
		}
		// some images on the Hub have empty tags - treat those as "latest"
		set := sets.NewString(tags...)
		if set.Has("") {
			set.Delete("")
			set.Insert(api.DefaultImageTag)
		}
		tags = set.List()
		// include only the top N tags in the result, put the rest in AdditionalTags
		api.PrioritizeTags(tags)
		for _, s := range tags {
			if count <= 0 {
				repository.AdditionalTags = append(repository.AdditionalTags, s)
				continue
			}
			count--
			repository.Tags = append(repository.Tags, importTag{
				Name: s,
			})
		}
	}

	// load digests
	for i := range repository.Digests {
		importDigest := &repository.Digests[i]
		if importDigest.Err != nil || importDigest.Image != nil {
			continue
		}
		limiter.Accept()
		image, err := conn.ImageByID(repository.Ref.Namespace, repository.Ref.Name, importDigest.Name)
		if err != nil {
			importDigest.Err = err
			continue
		}
		// we do not preserve manifests of legacy images
		importDigest.Image, err = schema0ToImage(image, importDigest.Name)
		if err != nil {
			importDigest.Err = err
			continue
		}
	}

	for i := range repository.Tags {
		importTag := &repository.Tags[i]
		if importTag.Err != nil || importTag.Image != nil {
			continue
		}
		limiter.Accept()
		image, err := conn.ImageByTag(repository.Ref.Namespace, repository.Ref.Name, importTag.Name)
		if err != nil {
			importTag.Err = err
			continue
		}
		// we do not preserve manifests of legacy images
		importTag.Image, err = schema0ToImage(image, "")
		if err != nil {
			importTag.Err = err
			continue
		}
	}
}

type importTag struct {
	Name  string
	Image *api.Image
	Err   error
}

type importDigest struct {
	Name  string
	Image *api.Image
	Err   error
}

type importRepository struct {
	Ref      api.DockerImageReference
	Registry *url.URL
	Name     string
	Insecure bool

	Tags    []importTag
	Digests []importDigest

	MaximumTags    int
	AdditionalTags []string
	Err            error
}

// repositoryKey is the key used to cache information loaded from a remote Docker repository.
type repositoryKey struct {
	// The URL of the server
	url url.URL
	// The name of the image repository (contains both namespace and path)
	name string
}

// manifestKey is a key for a map between a Docker image tag or image ID and a retrieved api.Image, used
// to ensure we don't fetch the same image multiple times.
type manifestKey struct {
	repositoryKey
	// The tag or ID of the image, not used within the same map
	value string
}

func imageImportStatus(err error, kind, position string) unversioned.Status {
	switch t := err.(type) {
	case kapierrors.APIStatus:
		return t.Status()
	case *field.Error:
		return kapierrors.NewInvalid(api.Kind(kind), position, field.ErrorList{t}).(kapierrors.APIStatus).Status()
	default:
		return kapierrors.NewInternalError(err).(kapierrors.APIStatus).Status()
	}
}

func setImageImportStatus(images *api.ImageStreamImport, i int, err error) {
	images.Status.Images[i].Status = imageImportStatus(err, "", "")
}

func invalidStatus(position string, errs ...*field.Error) unversioned.Status {
	return kapierrors.NewInvalid(api.Kind(""), position, errs).(kapierrors.APIStatus).Status()
}

// NewContext is capable of creating RepositoryRetrievers.
func NewContext(transport, insecureTransport http.RoundTripper) Context {
	return Context{
		Transport:         transport,
		InsecureTransport: insecureTransport,
		Challenges:        auth.NewSimpleChallengeManager(),
	}
}

type Context struct {
	Transport         http.RoundTripper
	InsecureTransport http.RoundTripper
	Challenges        auth.ChallengeManager
}

func (c Context) WithCredentials(credentials auth.CredentialStore) RepositoryRetriever {
	return &repositoryRetriever{
		context:     c,
		credentials: credentials,

		pings:    make(map[url.URL]error),
		redirect: make(map[url.URL]*url.URL),
	}
}

type repositoryRetriever struct {
	context     Context
	credentials auth.CredentialStore

	pings    map[url.URL]error
	redirect map[url.URL]*url.URL
}

func (r *repositoryRetriever) Repository(ctx gocontext.Context, registry *url.URL, repoName string, insecure bool) (distribution.Repository, error) {
	t := r.context.Transport
	if insecure && r.context.InsecureTransport != nil {
		t = r.context.InsecureTransport
	}
	src := *registry
	// ping the registry to get challenge headers
	if err, ok := r.pings[src]; ok {
		if err != nil {
			return nil, err
		}
		if redirect, ok := r.redirect[src]; ok {
			src = *redirect
		}
	} else {
		redirect, err := r.ping(src, insecure, t)
		r.pings[src] = err
		if err != nil {
			return nil, err
		}
		if redirect != nil {
			r.redirect[src] = redirect
			src = *redirect
		}
	}

	rt := transport.NewTransport(
		t,
		// TODO: slightly smarter authorizer that retries unauthenticated requests
		// TODO: make multiple attempts if the first credential fails
		auth.NewAuthorizer(
			r.context.Challenges,
			auth.NewTokenHandler(t, r.credentials, repoName, "pull"),
			auth.NewBasicHandler(r.credentials),
		),
	)
	return registryclient.NewRepository(context.Context(ctx), repoName, src.String(), rt)
}

func (r *repositoryRetriever) ping(registry url.URL, insecure bool, transport http.RoundTripper) (*url.URL, error) {
	pingClient := &http.Client{
		Transport: transport,
		Timeout:   15 * time.Second,
	}
	target := registry
	target.Path = path.Join(target.Path, "v2") + "/"
	req, err := http.NewRequest("GET", target.String(), nil)
	if err != nil {
		return nil, err
	}
	resp, err := pingClient.Do(req)
	if err != nil {
		if insecure && registry.Scheme == "https" {
			glog.V(5).Infof("Falling back to an HTTP check for an insecure registry %s: %v", registry, err)
			registry.Scheme = "http"
			_, nErr := r.ping(registry, true, transport)
			if nErr != nil {
				return nil, nErr
			}
			return &registry, nil
		}
		return nil, err
	}
	defer resp.Body.Close()

	versions := auth.APIVersions(resp, "Docker-Distribution-API-Version")
	if len(versions) == 0 {
		glog.V(5).Infof("Registry responded to v2 Docker endpoint, but has no header for Docker Distribution %s: %d, %#v", req.URL, resp.StatusCode, resp.Header)
		return nil, &ErrNotV2Registry{Registry: registry.String()}
	}

	r.context.Challenges.AddResponse(resp)

	return nil, nil
}

func schema1ToImage(manifest *schema1.SignedManifest, d digest.Digest) (*api.Image, error) {
	if len(manifest.History) == 0 {
		return nil, fmt.Errorf("image has no v1Compatibility history and cannot be used")
	}
	dockerImage, err := unmarshalDockerImage([]byte(manifest.History[0].V1Compatibility))
	if err != nil {
		return nil, err
	}
	if len(d) > 0 {
		dockerImage.ID = d.String()
	} else {
		if p, err := manifest.Payload(); err == nil {
			d, err := digest.FromBytes(p)
			if err != nil {
				return nil, fmt.Errorf("unable to create digest from image payload: %v", err)
			}
			dockerImage.ID = d.String()
		} else {
			d, err := digest.FromBytes(manifest.Raw)
			if err != nil {
				return nil, fmt.Errorf("unable to create digest from image bytes: %v", err)
			}
			dockerImage.ID = d.String()
		}
	}
	image := &api.Image{
		ObjectMeta: kapi.ObjectMeta{
			Name: dockerImage.ID,
		},
		DockerImageMetadata:        *dockerImage,
		DockerImageManifest:        string(manifest.Raw),
		DockerImageMetadataVersion: "1.0",
	}

	return image, nil
}

func schema0ToImage(dockerImage *dockerregistry.Image, id string) (*api.Image, error) {
	var baseImage api.DockerImage
	if err := kapi.Scheme.Convert(&dockerImage.Image, &baseImage); err != nil {
		return nil, fmt.Errorf("could not convert image: %#v", err)
	}

	image := &api.Image{
		ObjectMeta: kapi.ObjectMeta{
			Name: dockerImage.ID,
		},
		DockerImageMetadata:        baseImage,
		DockerImageMetadataVersion: "1.0",
	}

	return image, nil
}

func unmarshalDockerImage(body []byte) (*api.DockerImage, error) {
	var image dockerpre012.DockerImage
	if err := json.Unmarshal(body, &image); err != nil {
		return nil, err
	}
	dockerImage := &api.DockerImage{}
	if err := kapi.Scheme.Convert(&image, dockerImage); err != nil {
		return nil, err
	}
	return dockerImage, nil
}

func isDockerError(err error, code errcode.ErrorCode) bool {
	switch t := err.(type) {
	case errcode.Errors:
		for _, err := range t {
			if isDockerError(err, code) {
				return true
			}
		}
	case errcode.ErrorCode:
		if code == t {
			return true
		}
	case errcode.Error:
		if t.ErrorCode() == code {
			return true
		}
	}
	return false
}

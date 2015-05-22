
package io.fabric8.openshift.api.model.template;

import com.fasterxml.jackson.annotation.*;
import io.fabric8.KubernetesJson;
import io.fabric8.kubernetes.api.model.*;
import io.fabric8.openshift.api.model.*;
import org.apache.commons.lang.builder.EqualsBuilder;
import org.apache.commons.lang.builder.HashCodeBuilder;
import org.apache.commons.lang.builder.ToStringBuilder;

import javax.annotation.Generated;
import javax.validation.Valid;
import javax.validation.constraints.NotNull;
import java.util.*;


/**
 *
 *
 */
@JsonInclude(JsonInclude.Include.NON_NULL)
@Generated("org.jsonschema2pojo")
@JsonPropertyOrder({
    "apiVersion",
    "kind",
    "labels",
    "metadata",
    "objects",
    "parameters"
})
public class Template implements KubernetesJson, HasMetadata {

    /**
     *
     * (Required)
     *
     */
    @JsonProperty("apiVersion")
    @NotNull
    private Template.ApiVersion apiVersion = Template.ApiVersion.fromValue("v1beta3");
    /**
     *
     * (Required)
     *
     */
    @JsonProperty("kind")
    @NotNull
    private java.lang.String kind = "Template";
    /**
     *
     *
     */
    @JsonProperty("labels")
    @Valid
    private Map<String, String> labels;
    /**
     *
     *
     */
    @JsonProperty("metadata")
    @Valid
    private ObjectMeta metadata;
    /**
     *
     *
     */
    @JsonProperty("objects")
    @Valid
    private List<HasMetadata> objects = new ArrayList<HasMetadata>();
    /**
     *
     *
     */
    @JsonProperty("parameters")
    @Valid
    private List<Parameter> parameters = new ArrayList<Parameter>();
    @JsonIgnore
    private Map<java.lang.String, java.lang.Object> additionalProperties = new HashMap<java.lang.String, java.lang.Object>();

    /**
     * list of services
     * Note: This is not to be used. Added for influencing the generation of fluent nested builders
     */
    @JsonIgnore
    private final List<Service> services = new ArrayList<>();

    /**
     * list of replication controllers
     * Note: This is not to be used. Added for influencing the generation of fluent nested builders
     */
    @JsonIgnore
    private final List<ReplicationController> replicationControllers = new ArrayList<>();

    /**
     * list of pods.
     * Note: This is not to be used. Added for influencing the generation of fluent nested builders
     *
     */
    @JsonIgnore
    private final List<Pod> pods = new ArrayList<>();

    /**
     * list of build configs
     * Note: This is not to be used. Added for influencing the generation of fluent nested builders
     *
     */
    @JsonIgnore
    private final List<BuildConfig> buildConfigs = new ArrayList<>();

    /**
     * list of deployment configs
     * Note: This is not to be used. Added for influencing the generation of fluent nested builders
     */
    @JsonIgnore
    private final List<DeploymentConfig> deploymentConfigs = new ArrayList<>();

    /**
     * list of image repositories
     * Note: This is not to be used. Added for influencing the generation of fluent nested builders
     */
    @JsonIgnore
    private final List<ImageStream> imageStreams = new ArrayList<>();

    /**
     * list of routes
     * Note: This is not to be used. Added for influencing the generation of fluent nested builders.
     */
    @JsonIgnore
    private final List<Route> routes = new ArrayList<>();

    /**
     * list of routes
     * Note: This is not to be used. Added for influencing the generation of fluent nested builders.
     */
    @JsonIgnore
    private final List<Template> templates = new ArrayList<>();

    /**
     * list of oauth clients
     * Note: This is not to be used. Added for influencing the generation of fluent nested builders.
     */
    @JsonIgnore
    private final List<OAuthClient> oAuthClients = new ArrayList<>();

    /**
     * list of oauth client authorizations
     * Note: This is not to be used. Added for influencing the generation of fluent nested builders.
     */
    @JsonIgnore
    private final List<OAuthClientAuthorization> oAuthClientAuthorizations = new ArrayList<>();

    /**
     * list of oauth access tokens
     * Note: This is not to be used. Added for influencing the generation of fluent nested builders.
     */
    @JsonIgnore
    private final List<OAuthAccessToken> oAuthAccessTokens = new ArrayList<>();

    /**
     * list of namespaces
     * Note: This is not to be used. Added for influencing the generation of fluent nested builders.
     */
    @JsonIgnore
    private final List<Namespace> namespaces = new ArrayList<>();

    /**
     * list of secretes
     * Note: This is not to be used. Added for influencing the generation of fluent nested builders.
     */
    @JsonIgnore
    private final List<Secret> secrets = new ArrayList<>();

    /**
     * No args constructor for use in serialization
     *
     */
    public Template() {
    }

    /**
     *
     * @param apiVersion
     * @param labels
     * @param parameters
     * @param objects
     * @param kind
     * @param metadata
     */
    public Template(Template.ApiVersion apiVersion, java.lang.String kind, Map<String, String> labels, ObjectMeta metadata, List<HasMetadata> objects, List<Parameter> parameters,List<Service> services,
                    List<ReplicationController> replicationControllers,
                    List<Pod> pods,
                    List<BuildConfig> buildConfigs,
                    List<DeploymentConfig> deploymentConfigs,
                    List<ImageStream> imageStreams,
                    List<Route> routes,
                    List<Template> templates,
                    List<OAuthClient> oAuthClients,
                    List<OAuthClientAuthorization> oAuthClientAuthorizations,
                    List<OAuthAccessToken> oAuthAccessTokens,
                    List<Namespace> namespaces,
                    List<Secret> secrets) {
        this.apiVersion = apiVersion;
        this.kind = kind;
        this.labels = labels;
        this.metadata = metadata;
        this.objects = objects;
        this.parameters = parameters;

        this.setObjects(objects);
        this.services.addAll(services != null ? services : Collections.<Service>emptyList());
        this.replicationControllers.addAll(replicationControllers != null ? replicationControllers : Collections.<ReplicationController>emptyList());
        this.pods.addAll(pods != null ? pods : Collections.<Pod>emptyList());
        this.buildConfigs.addAll(buildConfigs != null ? buildConfigs : Collections.<BuildConfig>emptyList());
        this.deploymentConfigs.addAll(deploymentConfigs != null ? deploymentConfigs : Collections.<DeploymentConfig>emptyList());
        this.imageStreams.addAll(imageStreams != null ? imageStreams : Collections.<ImageStream>emptyList());
        this.routes.addAll(routes != null ? routes : Collections.<Route>emptyList());
        this.templates.addAll(templates != null ? templates : Collections.<Template>emptyList());
        this.oAuthClients.addAll(oAuthClients != null ? oAuthClients : Collections.<OAuthClient>emptyList());
        this.oAuthClientAuthorizations.addAll(oAuthClientAuthorizations != null ? oAuthClientAuthorizations : Collections.<OAuthClientAuthorization>emptyList());
        this.oAuthAccessTokens.addAll(oAuthAccessTokens != null ? oAuthAccessTokens : Collections.<OAuthAccessToken>emptyList());
        this.namespaces.addAll(namespaces != null ? namespaces : Collections.<Namespace>emptyList());
        this.secrets.addAll(secrets != null ? secrets : Collections.<Secret>emptyList());
    }

    /**
     *
     * (Required)
     *
     * @return
     *     The apiVersion
     */
    @JsonProperty("apiVersion")
    public Template.ApiVersion getApiVersion() {
        return apiVersion;
    }

    /**
     *
     * (Required)
     *
     * @param apiVersion
     *     The apiVersion
     */
    @JsonProperty("apiVersion")
    public void setApiVersion(Template.ApiVersion apiVersion) {
        this.apiVersion = apiVersion;
    }

    /**
     *
     * (Required)
     *
     * @return
     *     The kind
     */
    @JsonProperty("kind")
    public java.lang.String getKind() {
        return kind;
    }

    /**
     *
     * (Required)
     *
     * @param kind
     *     The kind
     */
    @JsonProperty("kind")
    public void setKind(java.lang.String kind) {
        this.kind = kind;
    }

    /**
     *
     *
     * @return
     *     The labels
     */
    @JsonProperty("labels")
    public Map<String, String> getLabels() {
        return labels;
    }

    /**
     *
     *
     * @param labels
     *     The labels
     */
    @JsonProperty("labels")
    public void setLabels(Map<String, String> labels) {
        this.labels = labels;
    }

    /**
     *
     *
     * @return
     *     The metadata
     */
    @JsonProperty("metadata")
    public ObjectMeta getMetadata() {
        return metadata;
    }

    /**
     *
     *
     * @param metadata
     *     The metadata
     */
    @JsonProperty("metadata")
    public void setMetadata(ObjectMeta metadata) {
        this.metadata = metadata;
    }

    /**
     *
     *
     * @return
     *     The objects
     */
    @JsonProperty("objects")
    public List<HasMetadata> getObjects() {
        HasMetadatSet allItems = new HasMetadatSet(objects);
        allItems.addAll(services);
        allItems.addAll(replicationControllers);
        allItems.addAll(pods);
        allItems.addAll(buildConfigs);
        allItems.addAll(deploymentConfigs);
        allItems.addAll(imageStreams);
        allItems.addAll(routes);
        allItems.addAll(templates);
        allItems.addAll(oAuthClients);
        allItems.addAll(oAuthClientAuthorizations);
        allItems.addAll(oAuthAccessTokens);
        allItems.addAll(namespaces);
        allItems.addAll(secrets);
        return new ArrayList<>(allItems);
    }

    /**
     *
     *
     * @param objects
     *     The objects
     */
    @JsonProperty("objects")
    @JsonTypeInfo(
            use = JsonTypeInfo.Id.NAME,
            include = JsonTypeInfo.As.PROPERTY,
            property = "kind")
    @JsonSubTypes({
            @JsonSubTypes.Type(value = KubernetesList.class, name = "List"),
            @JsonSubTypes.Type(value = Service.class, name = "Service"),
            @JsonSubTypes.Type(value = Pod.class, name = "Pod"),
            @JsonSubTypes.Type(value = ReplicationController.class, name = "ReplicationController"),
            @JsonSubTypes.Type(value = BuildConfig.class, name = "BuildConfig"),
            @JsonSubTypes.Type(value = DeploymentConfig.class, name = "DeploymentConfig"),
            @JsonSubTypes.Type(value = ImageStream.class, name = "ImageStream"),
            @JsonSubTypes.Type(value = Route.class, name = "Route"),
            @JsonSubTypes.Type(value = Template.class, name = "Template"),
            @JsonSubTypes.Type(value = OAuthClient.class, name = "OAuthClient"),
            @JsonSubTypes.Type(value = OAuthClientAuthorization.class, name = "OAuthClientAuthorization"),
            @JsonSubTypes.Type(value = OAuthAccessToken.class, name = "OAuthAccessToken"),
            @JsonSubTypes.Type(value = Namespace.class, name = "Namespace"),
            @JsonSubTypes.Type(value = Secret.class, name = "Secrets")

    })
    public void setObjects(List<HasMetadata> objects) {
        this.services.clear();
        this.replicationControllers.clear();
        this.pods.clear();
        this.buildConfigs.clear();
        this.deploymentConfigs.clear();
        this.imageStreams.clear();
        this.routes.clear();
        this.templates.clear();
        this.oAuthClients.clear();
        this.oAuthClientAuthorizations.clear();
        this.oAuthAccessTokens.clear();
        this.namespaces.clear();
        this.secrets.clear();

        for (HasMetadata item : objects) {
            if (item instanceof Service) {
                this.services.add((Service) item);
            } else if (item instanceof ReplicationController) {
                this.replicationControllers.add((ReplicationController) item);
            } else if (item instanceof Pod) {
                this.pods.add((Pod) item);
            } else if (item instanceof BuildConfig) {
                this.buildConfigs.add((BuildConfig) item);
            } else if (item instanceof DeploymentConfig) {
                this.deploymentConfigs.add((DeploymentConfig) item);
            } else if (item instanceof ImageStream) {
                this.imageStreams.add((ImageStream) item);
            } else if (item instanceof Route) {
                this.routes.add((Route) item);
            } else if (item instanceof Template) {
                this.templates.add((Template) item);
            }
        }
    }

    /**
     *
     *
     * @return
     *     The parameters
     */
    @JsonProperty("parameters")
    public List<Parameter> getParameters() {
        return parameters;
    }

    /**
     *
     *
     * @param parameters
     *     The parameters
     */
    @JsonProperty("parameters")
    public void setParameters(List<Parameter> parameters) {
        this.parameters = parameters;
    }

    @Override
    public java.lang.String toString() {
        return ToStringBuilder.reflectionToString(this);
    }

    @JsonAnyGetter
    public Map<java.lang.String, java.lang.Object> getAdditionalProperties() {
        return this.additionalProperties;
    }

    @JsonAnySetter
    public void setAdditionalProperty(java.lang.String name, java.lang.Object value) {
        this.additionalProperties.put(name, value);
    }

    @Override
    public int hashCode() {
        return new HashCodeBuilder().append(apiVersion).append(kind).append(labels).append(metadata).append(objects).append(parameters).append(additionalProperties).toHashCode();
    }

    @Override
    public boolean equals(java.lang.Object other) {
        if (other == this) {
            return true;
        }
        if ((other instanceof Template) == false) {
            return false;
        }
        Template rhs = ((Template) other);
        return new EqualsBuilder().append(apiVersion, rhs.apiVersion).append(kind, rhs.kind).append(labels, rhs.labels).append(metadata, rhs.metadata).append(objects, rhs.objects).append(parameters, rhs.parameters).append(additionalProperties, rhs.additionalProperties).isEquals();
    }

    @Generated("org.jsonschema2pojo")
    public static enum ApiVersion {

        V_1_BETA_1("v1beta1"),
        V_1_BETA_2("v1beta2"),
        V_1_BETA_3("v1beta3");
        private final java.lang.String value;
        private static Map<java.lang.String, Template.ApiVersion> constants = new HashMap<java.lang.String, Template.ApiVersion>();

        static {
            for (Template.ApiVersion c: values()) {
                constants.put(c.value, c);
            }
        }

        private ApiVersion(java.lang.String value) {
            this.value = value;
        }

        @JsonValue
        @Override
        public java.lang.String toString() {
            return this.value;
        }

        @JsonCreator
        public static Template.ApiVersion fromValue(java.lang.String value) {
            Template.ApiVersion constant = constants.get(value);
            if (constant == null) {
                throw new IllegalArgumentException(value);
            } else {
                return constant;
            }
        }

    }


    @JsonIgnore
    public List<Service> getServices() {
        return services;
    }

    @JsonIgnore
    public List<ReplicationController> getReplicationControllers() {
        return replicationControllers;
    }

    @JsonIgnore
    public List<Pod> getPods() {
        return pods;
    }

    @JsonIgnore
    public List<BuildConfig> getBuildConfigs() {
        return buildConfigs;
    }

    @JsonIgnore
    public List<DeploymentConfig> getDeploymentConfigs() {
        return deploymentConfigs;
    }

    @JsonIgnore
    public List<ImageStream> getImageStreams() {
        return imageStreams;
    }

    @JsonIgnore
    public List<Route> getRoutes() {
        return routes;
    }

    @JsonIgnore
    public List<Template> getTemplates() {
        return templates;
    }

    @JsonIgnore
    public List<OAuthClient> getOAuthClients() {
        return oAuthClients;
    }

    @JsonIgnore
    public List<OAuthClientAuthorization> getOAuthClientAuthorizations() {
        return oAuthClientAuthorizations;
    }

    @JsonIgnore
    public List<OAuthAccessToken> getOAuthAccessTokens() {
        return oAuthAccessTokens;
    }

    @JsonIgnore
    public List<Namespace> getNamespaces() {
        return namespaces;
    }

    @JsonIgnore
    public List<Secret> getSecrets() {
        return secrets;
    }
}

package servicefabricmesh

import (
	"fmt"
	"log"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/preview/servicefabricmesh/mgmt/2018-09-01-preview/servicefabricmesh"
	"github.com/hashicorp/go-azure-helpers/lang/response"
	"github.com/hashicorp/go-azure-helpers/resourcemanager/location"
	"github.com/hashicorp/terraform-provider-azurerm/helpers/azure"
	"github.com/hashicorp/terraform-provider-azurerm/helpers/tf"
	"github.com/hashicorp/terraform-provider-azurerm/internal/clients"
	"github.com/hashicorp/terraform-provider-azurerm/internal/services/servicefabricmesh/parse"
	"github.com/hashicorp/terraform-provider-azurerm/internal/tags"
	"github.com/hashicorp/terraform-provider-azurerm/internal/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/internal/tf/suppress"
	"github.com/hashicorp/terraform-provider-azurerm/internal/tf/validation"
	"github.com/hashicorp/terraform-provider-azurerm/internal/timeouts"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

func resourceServiceFabricMeshSecretValue() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Create: resourceServiceFabricMeshSecretValueCreateUpdate,
		Read:   resourceServiceFabricMeshSecretValueRead,
		Update: resourceServiceFabricMeshSecretValueCreateUpdate,
		Delete: resourceServiceFabricMeshSecretValueDelete,
		Importer: pluginsdk.ImporterValidatingResourceId(func(id string) error {
			_, err := parse.SecretValueID(id)
			return err
		}),

		DeprecationMessage: deprecationMessage("azurerm_service_fabric_mesh_secret_value"),

		Timeouts: &pluginsdk.ResourceTimeout{
			Create: pluginsdk.DefaultTimeout(30 * time.Minute),
			Read:   pluginsdk.DefaultTimeout(5 * time.Minute),
			Update: pluginsdk.DefaultTimeout(30 * time.Minute),
			Delete: pluginsdk.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*pluginsdk.Schema{
			"name": {
				Type:     pluginsdk.TypeString,
				Required: true,
				ForceNew: true,
				// Follow casing issue here https://github.com/Azure/azure-rest-api-specs/issues/9330
				DiffSuppressFunc: suppress.CaseDifference,
				ValidateFunc:     validation.StringIsNotEmpty,
			},

			"service_fabric_mesh_secret_id": {
				Type:     pluginsdk.TypeString,
				Required: true,
				ForceNew: true,
				// Follow casing issue here https://github.com/Azure/azure-rest-api-specs/issues/9330
				DiffSuppressFunc: suppress.CaseDifference,
				ValidateFunc:     azure.ValidateResourceID,
			},

			"location": azure.SchemaLocation(),

			"value": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},

			"tags": tags.Schema(),
		},
	}
}

func resourceServiceFabricMeshSecretValueCreateUpdate(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).ServiceFabricMesh.SecretValueClient
	ctx, cancel := timeouts.ForCreateUpdate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	location := location.Normalize(d.Get("location").(string))
	t := d.Get("tags").(map[string]interface{})

	secretID, err := parse.SecretID(d.Get("service_fabric_mesh_secret_id").(string))
	if err != nil {
		return err
	}
	id := parse.NewSecretValueID(secretID.SubscriptionId, secretID.ResourceGroup, secretID.Name, d.Get("name").(string))

	if d.IsNewResource() {
		existing, err := client.Get(ctx, id.ResourceGroup, id.SecretName, id.ValueName)
		if err != nil {
			if !utils.ResponseWasNotFound(existing.Response) {
				return fmt.Errorf("checking for presence of existing Service Fabric Mesh Secret Value: %+v", err)
			}
		}

		if existing.ID != nil && *existing.ID != "" {
			return tf.ImportAsExistsError("azurerm_service_fabric_mesh_secret_value", *existing.ID)
		}
	}

	parameters := servicefabricmesh.SecretValueResourceDescription{
		SecretValueResourceProperties: &servicefabricmesh.SecretValueResourceProperties{
			Value: utils.String(d.Get("value").(string)),
		},
		Location: utils.String(location),
		Tags:     tags.Expand(t),
	}

	if _, err := client.Create(ctx, id.ResourceGroup, id.SecretName, id.ValueName, parameters); err != nil {
		return fmt.Errorf("creating %s: %+v", id, err)
	}

	d.SetId(id.ID())

	return resourceServiceFabricMeshSecretValueRead(d, meta)
}

func resourceServiceFabricMeshSecretValueRead(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).ServiceFabricMesh.SecretValueClient
	secretClient := meta.(*clients.Client).ServiceFabricMesh.SecretClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.SecretValueID(d.Id())
	if err != nil {
		return err
	}

	secret, err := secretClient.Get(ctx, id.ResourceGroup, id.SecretName)
	if err != nil {
		if utils.ResponseWasNotFound(secret.Response) {
			log.Printf("[INFO] Unable to find Service Fabric Mesh Secret %q - removing from state", d.Id())
			d.SetId("")
			return nil
		}

		return fmt.Errorf("reading Service Fabric Mesh Secret: %+v", err)
	}

	resp, err := client.Get(ctx, id.ResourceGroup, id.SecretName, id.ValueName)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			log.Printf("[INFO] Unable to find Service Fabric Mesh Secret Value %q - removing from state", d.Id())
			d.SetId("")
			return nil
		}

		return fmt.Errorf("reading Service Fabric Mesh Secret Value: %+v", err)
	}

	d.Set("name", id.ValueName)
	d.Set("service_fabric_mesh_secret_id", secret.ID)
	d.Set("location", location.NormalizeNilable(resp.Location))

	return tags.FlattenAndSet(d, resp.Tags)
}

func resourceServiceFabricMeshSecretValueDelete(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).ServiceFabricMesh.SecretValueClient
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.SecretValueID(d.Id())
	if err != nil {
		return err
	}

	resp, err := client.Delete(ctx, id.ResourceGroup, id.SecretName, id.ValueName)
	if err != nil {
		if !response.WasNotFound(resp.Response) {
			return fmt.Errorf("deleting Service Fabric Mesh Secret Value %q (Resource Group %q / Secret %q): %+v", id.ValueName, id.ResourceGroup, id.SecretName, err)
		}
	}

	return nil
}

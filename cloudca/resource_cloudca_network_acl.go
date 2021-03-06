package cloudca

import (
	"fmt"
	"github.com/cloud-ca/go-cloudca"
	"github.com/cloud-ca/go-cloudca/api"
	"github.com/cloud-ca/go-cloudca/services/cloudca"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceCloudcaNetworkAcl() *schema.Resource {
	return &schema.Resource{
		Create: resourceCloudcaNetworkAclCreate,
		Read:   resourceCloudcaNetworkAclRead,
		Delete: resourceCloudcaNetworkAclDelete,

		Schema: map[string]*schema.Schema{
			"service_code": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "A cloudca service code",
			},
			"environment_name": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Name of environment where the network ACL should be created",
			},
			"name": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Name of network ACL",
			},
			"description": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Description of network ACL",
			},
			"vpc_id": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Id of the VPC",
			},
		},
	}
}

func resourceCloudcaNetworkAclCreate(d *schema.ResourceData, meta interface{}) error {
	ccaClient := meta.(*cca.CcaClient)
	resources, _ := ccaClient.GetResources(d.Get("service_code").(string), d.Get("environment_name").(string))
	ccaResources := resources.(cloudca.Resources)

	aclToCreate := cloudca.NetworkAcl{
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
		VpcId:       d.Get("vpc_id").(string),
	}
	newAcl, err := ccaResources.NetworkAcls.Create(aclToCreate)
	if err != nil {
		return fmt.Errorf("Error creating the new network ACL %s: %s", aclToCreate.Name, err)
	}
	d.SetId(newAcl.Id)
	return resourceCloudcaNetworkAclRead(d, meta)
}

func resourceCloudcaNetworkAclRead(d *schema.ResourceData, meta interface{}) error {
	ccaClient := meta.(*cca.CcaClient)
	resources, _ := ccaClient.GetResources(d.Get("service_code").(string), d.Get("environment_name").(string))
	ccaResources := resources.(cloudca.Resources)

	acl, aErr := ccaResources.NetworkAcls.Get(d.Id())
	if aErr != nil {
		if ccaError, ok := aErr.(api.CcaErrorResponse); ok {
			if ccaError.StatusCode == 404 {
				fmt.Errorf("ACL %s not found", d.Id())
				d.SetId("")
				return nil
			}
		}
		return aErr
	}

	// Update the config
	d.Set("name", acl.Name)
	d.Set("description", acl.Description)
	d.Set("vpc_id", acl.VpcId)

	return nil
}

func resourceCloudcaNetworkAclDelete(d *schema.ResourceData, meta interface{}) error {
	ccaClient := meta.(*cca.CcaClient)
	resources, _ := ccaClient.GetResources(d.Get("service_code").(string), d.Get("environment_name").(string))
	ccaResources := resources.(cloudca.Resources)
	if _, err := ccaResources.NetworkAcls.Delete(d.Id()); err != nil {
		if ccaError, ok := err.(api.CcaErrorResponse); ok {
			if ccaError.StatusCode == 404 {
				fmt.Errorf("Network ACL %s not found", d.Id())
				d.SetId("")
				return nil
			}
		}
		return err
	}
	return nil
}

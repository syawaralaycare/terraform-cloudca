package cloudca

import (
	"fmt"
	"github.com/cloud-ca/go-cloudca"
	"github.com/cloud-ca/go-cloudca/api"
	"github.com/cloud-ca/go-cloudca/services/cloudca"
	"github.com/hashicorp/terraform/helper/schema"
	"log"
)

func resourceCloudcaPublicIp() *schema.Resource {
	return &schema.Resource{
		Create: resourceCloudcaPublicIpCreate,
		Read:   resourceCloudcaPublicIpRead,
		Delete: resourceCloudcaPublicIpDelete,

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
				Description: "Name of environment where the public IP should be created",
			},
			"vpc_id": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Id of the VPC",
			},
			"ip_address": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceCloudcaPublicIpCreate(d *schema.ResourceData, meta interface{}) error {
	ccaClient := meta.(*cca.CcaClient)
	resources, _ := ccaClient.GetResources(d.Get("service_code").(string), d.Get("environment_name").(string))
	ccaResources := resources.(cloudca.Resources)

	vpcId := d.Get("vpc_id").(string)

	publicIpToCreate := cloudca.PublicIp{
		VpcId: vpcId,
	}
	newPublicIp, err := ccaResources.PublicIps.Acquire(publicIpToCreate)
	if err != nil {
		return fmt.Errorf("Error acquiring the new public ip %s", err)
	}
	d.SetId(newPublicIp.Id)
	return resourceCloudcaPublicIpRead(d, meta)
}

func resourceCloudcaPublicIpRead(d *schema.ResourceData, meta interface{}) error {
	ccaClient := meta.(*cca.CcaClient)
	resources, _ := ccaClient.GetResources(d.Get("service_code").(string), d.Get("environment_name").(string))
	ccaResources := resources.(cloudca.Resources)

	publicIp, err := ccaResources.PublicIps.Get(d.Id())
	if err != nil {
		if ccaError, ok := err.(api.CcaErrorResponse); ok {
			if ccaError.StatusCode == 404 {
				log.Printf("Public Ip with id='%s' was not found", d.Id())
				d.SetId("")
				return nil
			}
		}
		return err
	}
	d.Set("vpc_id", publicIp.VpcId)
	d.Set("ip_address", publicIp.IpAddress)
	return nil
}

func resourceCloudcaPublicIpDelete(d *schema.ResourceData, meta interface{}) error {
	ccaClient := meta.(*cca.CcaClient)
	resources, _ := ccaClient.GetResources(d.Get("service_code").(string), d.Get("environment_name").(string))
	ccaResources := resources.(cloudca.Resources)
	if _, err := ccaResources.PublicIps.Release(d.Id()); err != nil {
		if ccaError, ok := err.(api.CcaErrorResponse); ok {
			if ccaError.StatusCode == 404 {
				fmt.Errorf("Public Ip %s not found", d.Id())
				d.SetId("")
				return nil
			}
		}
		return err
	}

	return nil
}

package megaport

import (
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/utilitywarehouse/terraform-provider-megaport/megaport/api"
)

func resourceMegaportAwsVxc() *schema.Resource {
	return &schema.Resource{
		Create: resourceMegaportAwsVxcCreate,
		Read:   resourceMegaportAwsVxcRead,
		Update: resourceMegaportAwsVxcUpdate,
		Delete: resourceMegaportAwsVxcDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"rate_limit": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"a_end": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem:     resourceMegaportVxcEndElem(),
			},
			"b_end": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem:     resourceMegaportVxcAwsEndElem(),
			},
			"invoice_reference": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func resourceMegaportVxcAwsEndElem() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"product_uid": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"aws_account_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"customer_asn": {
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},
			"bgp_auth_key": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
		},
	}
}

func flattenVxcEndAws(v api.ProductAssociatedVxcEnd, r api.ProductAssociatedVxcResources) []interface{} {
	return []interface{}{map[string]interface{}{
		"product_uid":    v.ProductUid,
		"aws_account_id": r.AwsVirtualInterface.OwnerAccount,
		"customer_asn":   int(r.AwsVirtualInterface.Asn),
		"bgp_auth_key":   r.AwsVirtualInterface.AuthKey,
	}}
}

func resourceMegaportAwsVxcRead(d *schema.ResourceData, m interface{}) error {
	cfg := m.(*Config)
	p, err := cfg.Client.GetCloudVxc(d.Id())
	if err != nil {
		log.Printf("resourceMegaportAwsVxcRead: %v", err)
		d.SetId("")
		return nil
	}
	if p.ProvisioningStatus == api.ProductStatusDecommissioned {
		d.SetId("")
		return nil
	}
	if err := d.Set("name", p.ProductName); err != nil {
		return err
	}
	if err := d.Set("rate_limit", p.RateLimit); err != nil {
		return err
	}
	if err := d.Set("a_end", flattenVxcEnd(p.AEnd)); err != nil {
		return err
	}
	if err := d.Set("b_end", flattenVxcEndAws(p.BEnd, p.Resources)); err != nil {
		return err
	}
	if err := d.Set("invoice_reference", p.CostCentre); err != nil {
		return err
	}
	return nil
}

func resourceMegaportAwsVxcCreate(d *schema.ResourceData, m interface{}) error {
	cfg := m.(*Config)
	a := d.Get("a_end").([]interface{})[0].(map[string]interface{})
	b := d.Get("b_end").([]interface{})[0].(map[string]interface{})
	uid, err := cfg.Client.CreateCloudVxc(&api.CloudVxcCreateInput{
		ProductUidA:      api.String(a["product_uid"]),
		ProductUidB:      api.String(b["product_uid"]),
		Name:             api.String(d.Get("name")),
		InvoiceReference: api.String(d.Get("invoice_reference")),
		VlanA:            api.Uint64FromInt(a["vlan"]),
		RateLimit:        api.Uint64FromInt(d.Get("rate_limit")),
		PartnerConfig: &api.PartnerConfig{
			"connectType":       "AWS",
			"type":              "private",
			"asn":               b["customer_asn"],
			"ownerAccount":      b["aws_account_id"],
			"authKey":           b["bgp_auth_key"],
			"prefixes":          nil,
			"customerIpAddress": nil,
			"amazonIpAddress":   nil,
		},
	})
	if err != nil {
		return err
	}
	d.SetId(*uid)
	return resourceMegaportAwsVxcRead(d, m)
}

func resourceMegaportAwsVxcUpdate(d *schema.ResourceData, m interface{}) error {
	cfg := m.(*Config)
	a := d.Get("a_end").([]interface{})[0].(map[string]interface{})
	//b := d.Get("b_end").([]interface{})[0].(map[string]interface{})
	if err := cfg.Client.UpdateCloudVxc(&api.CloudVxcUpdateInput{
		InvoiceReference: api.String(d.Get("invoice_reference")),
		Name:             api.String(d.Get("name")),
		ProductUid:       api.String(d.Id()),
		RateLimit:        api.Uint64FromInt(d.Get("rate_limit")),
		VlanA:            api.Uint64FromInt(a["vlan"]),
	}); err != nil {
		return err
	}
	return resourceMegaportAwsVxcRead(d, m)
}

func resourceMegaportAwsVxcDelete(d *schema.ResourceData, m interface{}) error {
	cfg := m.(*Config)
	err := cfg.Client.DeleteCloudVxc(d.Id())
	if err != nil && err != api.ErrNotFound {
		return err
	}
	if err == api.ErrNotFound {
		log.Printf("resourceMegaportPortDelete: resource not found, deleting anyway")
	}
	return nil
}

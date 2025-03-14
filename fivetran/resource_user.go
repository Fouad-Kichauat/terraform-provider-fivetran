package fivetran

import (
	"context"
	"fmt"
	"time"

	"github.com/fivetran/go-fivetran"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceUser() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceUserCreate,
		ReadContext:   resourceUserRead,
		UpdateContext: resourceUserUpdate,
		DeleteContext: resourceUserDelete,
		Importer:      &schema.ResourceImporter{StateContext: schema.ImportStatePassthroughContext},
		Schema: map[string]*schema.Schema{
			"id": {Type: schema.TypeString, Computed: true},
			// The REST API doesn't provide a method to change the the user's email address.
			// That's why "ForceNew" is true.
			"email":       {Type: schema.TypeString, Required: true, ForceNew: true},
			"given_name":  {Type: schema.TypeString, Required: true},
			"family_name": {Type: schema.TypeString, Required: true},
			"role":        {Type: schema.TypeString, Required: true},

			"picture": {Type: schema.TypeString, Optional: true},
			"phone":   {Type: schema.TypeString, Optional: true},

			"logged_in_at": {Type: schema.TypeString, Computed: true},
			"created_at":   {Type: schema.TypeString, Computed: true},
			"last_updated": {Type: schema.TypeString, Computed: true}, // internal
			"verified":     {Type: schema.TypeBool, Computed: true},
			"invited":      {Type: schema.TypeBool, Computed: true},
		},
	}
}

func resourceUserCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*fivetran.Client)
	svc := client.NewUserInvite()

	svc.Email(d.Get("email").(string))
	svc.GivenName(d.Get("given_name").(string))
	svc.FamilyName(d.Get("family_name").(string))
	svc.Role(d.Get("role").(string))

	if v, ok := d.GetOk("picture"); ok && v != "" {
		svc.Picture(v.(string))
	}
	if v, ok := d.GetOk("phone"); ok && v != "" {
		svc.Phone(v.(string))
	}

	resp, err := svc.Do(ctx)
	if err != nil {
		return newDiagAppend(diags, diag.Error, "create error", fmt.Sprintf("%v; code: %v; message: %v", err, resp.Code, resp.Message))
	}

	d.SetId(resp.Data.ID)
	resourceUserRead(ctx, d, m)

	return diags
}

func resourceUserRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*fivetran.Client)
	svc := client.NewUserDetails()

	svc.UserID(d.Get("id").(string)).Do(ctx)

	resp, err := svc.Do(ctx)
	if err != nil {
		return newDiagAppend(diags, diag.Error, "read error", fmt.Sprintf("%v; code: %v; message: %v", err, resp.Code, resp.Message))
	}

	// msi stands for Map String Interface
	msi := make(map[string]interface{})
	msi["id"] = resp.Data.ID
	msi["email"] = resp.Data.Email
	msi["given_name"] = resp.Data.GivenName
	msi["family_name"] = resp.Data.FamilyName
	msi["verified"] = resp.Data.Verified
	msi["invited"] = resp.Data.Invited
	msi["picture"] = resp.Data.Picture
	msi["phone"] = resp.Data.Phone
	msi["logged_in_at"] = resp.Data.LoggedInAt.String()
	msi["created_at"] = resp.Data.CreatedAt.String()
	msi["role"] = resp.Data.Role
	for k, v := range msi {
		if err := d.Set(k, v); err != nil {
			return newDiagAppend(diags, diag.Error, "set error", fmt.Sprint(err))
		}
	}

	return diags
}

func resourceUserUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*fivetran.Client)
	svc := client.NewUserModify()

	svc.UserID(d.Get("id").(string))

	if d.HasChange("given_name") {
		svc.GivenName(d.Get("given_name").(string))
	}
	if d.HasChange("family_name") {
		svc.FamilyName(d.Get("family_name").(string))
	}
	if d.HasChange("picture") {
		svc.Picture(d.Get("picture").(string))
	}
	if d.HasChange("phone") {
		svc.Phone(d.Get("phone").(string))
	}
	if d.HasChange("role") {
		svc.Role(d.Get("role").(string))
	}

	resp, err := svc.Do(ctx)
	if err != nil {
		// resourceUserRead here makes sure the state is updated after a NewUserModify error.
		diags = resourceUserRead(ctx, d, m)
		return newDiagAppend(diags, diag.Error, "update error", fmt.Sprintf("%v; code: %v; message: %v", err, resp.Code, resp.Message))
	}

	if err := d.Set("last_updated", time.Now().Format(time.RFC850)); err != nil {
		return newDiagAppend(diags, diag.Error, "set error", fmt.Sprint(err))
	}

	return resourceUserRead(ctx, d, m)
}

func resourceUserDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*fivetran.Client)
	svc := client.NewUserDelete()

	resp, err := svc.UserID(d.Get("id").(string)).Do(ctx)
	if err != nil {
		return newDiagAppend(diags, diag.Error, "delete error", fmt.Sprintf("%v; code: %v; message: %v", err, resp.Code, resp.Message))
	}

	d.SetId("")

	return diags
}

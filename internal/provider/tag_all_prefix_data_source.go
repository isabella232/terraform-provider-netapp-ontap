package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/interfaces"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

// TODO:
// copy this file to match you data source (should match internal/provider/tag_prefix_data_source.go)
// replace GoAllPrefix with the name of the resource, following go conventions, eg IPInterfaces
// replace tag_all_prefix with the name of the resource, for logging purposes, eg ip_interfaces
// make sure to create internal/interfaces/tag_prefix.go too)
// delete these 5 lines

// Ensure provider defined types fully satisfy framework interfaces
var _ datasource.DataSource = &GoAllPrefixDataSource{}

// NewGoAllPrefixDataSource is a helper function to simplify the provider implementation.
func NewGoAllPrefixDataSource() datasource.DataSource {
	return &GoAllPrefixDataSource{
		config: resourceOrDataSourceConfig{
			name: "tag_all_prefix_data_source",
		},
	}
}

// GoAllPrefixDataSource defines the data source implementation.
type GoAllPrefixDataSource struct {
	config resourceOrDataSourceConfig
}

// GoAllPrefixDataSourceModel describes the data source data model.
type GoAllPrefixDataSourceModel struct {
	CxProfileName types.String                   `tfsdk:"cx_profile_name"`
	GoAllPrefix   []GoPrefixDataSourceModel      `tfsdk:"tag_all_prefix"`
	Filter        *GoPrefixDataSourceFilterModel `tfsdk:"filter"`
}

// Metadata returns the data source type name.
func (d *GoAllPrefixDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + d.config.name
}

// Schema defines the schema for the data source.
func (d *GoAllPrefixDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "GoAllPrefix data source",

		Attributes: map[string]schema.Attribute{
			"cx_profile_name": schema.StringAttribute{
				MarkdownDescription: "Connection profile name",
				Required:            true,
			},
			"filter": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"name": schema.StringAttribute{
						MarkdownDescription: "GoPrefix name",
						Optional:            true,
					},
					"svm_name": schema.StringAttribute{
						MarkdownDescription: "GoPrefix svm name",
						Optional:            true,
					},
				},
				Optional: true,
			},
			"tag_all_prefix": schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"cx_profile_name": schema.StringAttribute{
							MarkdownDescription: "Connection profile name",
							Required:            true,
						},
						"name": schema.StringAttribute{
							MarkdownDescription: "GoPrefix name",
							Required:            true,
						},
					},
				},
				Computed:            true,
				MarkdownDescription: "",
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *GoAllPrefixDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}
	config, ok := req.ProviderData.(Config)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected Config, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
	}
	d.config.providerConfig = config
}

// Read refreshes the Terraform state with the latest data.
func (d *GoAllPrefixDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data GoAllPrefixDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)
	// we need to defer setting the client until we can read the connection profile name
	client, err := getRestClient(errorHandler, d.config, data.CxProfileName)
	if err != nil {
		// error reporting done inside NewClient
		return
	}

	var filter *interfaces.GoPrefixGetDataModelONTAP = nil
	if data.Filter != nil {
		filter = &interfaces.GoPrefixGetDataModelONTAP{
			Name: data.Filter.Name.ValueString(),
		}
	}
	restInfo, err := interfaces.GetGoAllPrefix(errorHandler, *client, filter)
	if err != nil {
		// error reporting done inside GetGoAllPrefix
		return
	}

	data.GoAllPrefix = make([]GoPrefixDataSourceModel, len(restInfo))
	for index, record := range restInfo {
		data.GoAllPrefix[index] = GoPrefixDataSourceModel{
			CxProfileName: types.String(data.CxProfileName),
			Name:          types.StringValue(record.Name),
		}
	}

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Debug(ctx, fmt.Sprintf("read a data source: %#v", data))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

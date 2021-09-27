package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	tftags "github.com/hashicorp/terraform-provider-aws/aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func DataSourceTransitGatewayRouteTables() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceTransitGatewayRouteTablesRead,

		Schema: map[string]*schema.Schema{
			"filter": ec2CustomFiltersSchema(),

			"ids": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},

			"tags": tftags.TagsSchemaComputed(),
		},
	}
}

func dataSourceTransitGatewayRouteTablesRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	input := &ec2.DescribeTransitGatewayRouteTablesInput{}

	input.Filters = append(input.Filters, buildEC2TagFilterList(
		tftags.New(d.Get("tags").(map[string]interface{})).Ec2Tags(),
	)...)

	input.Filters = append(input.Filters, buildEC2CustomFilterList(
		d.Get("filter").(*schema.Set),
	)...)

	if len(input.Filters) == 0 {
		// Don't send an empty filters list; the EC2 API won't accept it.
		input.Filters = nil
	}

	var transitGatewayRouteTables []*ec2.TransitGatewayRouteTable

	err := conn.DescribeTransitGatewayRouteTablesPages(input, func(page *ec2.DescribeTransitGatewayRouteTablesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		transitGatewayRouteTables = append(transitGatewayRouteTables, page.TransitGatewayRouteTables...)

		return !lastPage
	})

	if err != nil {
		return fmt.Errorf("error describing EC2 Transit Gateway Route Tables: %w", err)
	}

	if len(transitGatewayRouteTables) == 0 {
		return fmt.Errorf("no matching EC2 Transit Gateway Route Tables found")
	}

	var ids []string

	for _, transitGatewayRouteTable := range transitGatewayRouteTables {
		if transitGatewayRouteTable == nil {
			continue
		}

		ids = append(ids, aws.StringValue(transitGatewayRouteTable.TransitGatewayRouteTableId))
	}

	d.SetId(meta.(*conns.AWSClient).Region)

	if err = d.Set("ids", ids); err != nil {
		return fmt.Errorf("error setting ids: %w", err)
	}

	return nil
}
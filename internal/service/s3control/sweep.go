// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build sweep
// +build sweep

package s3control

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3control"
	multierror "github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func init() {
	resource.AddTestSweepers("aws_s3_access_point", &resource.Sweeper{
		Name: "aws_s3_access_point",
		F:    sweepAccessPoints,
		Dependencies: []string{
			"aws_s3control_object_lambda_access_point",
		},
	})

	resource.AddTestSweepers("aws_s3control_multi_region_access_point", &resource.Sweeper{
		Name: "aws_s3control_multi_region_access_point",
		F:    sweepMultiRegionAccessPoints,
	})

	resource.AddTestSweepers("aws_s3control_object_lambda_access_point", &resource.Sweeper{
		Name: "aws_s3control_object_lambda_access_point",
		F:    sweepObjectLambdaAccessPoints,
	})

	resource.AddTestSweepers("aws_s3control_storage_lens_configuration", &resource.Sweeper{
		Name: "aws_s3control_storage_lens_configuration",
		F:    sweepStorageLensConfigurations,
	})
}

func sweepAccessPoints(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.S3ControlClient(ctx)
	accountID := client.AccountID
	input := &s3control.ListAccessPointsInput{
		AccountId: aws.String(accountID),
	}
	sweepResources := make([]sweep.Sweepable, 0)
	var sweeperErrs *multierror.Error

	pages := s3control.NewListAccessPointsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping S3 Access Point sweep for %s: %s", region, err)
			return sweeperErrs.ErrorOrNil()
		}

		if err != nil {
			return fmt.Errorf("error listing S3 Access Points (%s): %w", region, err)
		}

		for _, v := range page.AccessPointList {
			r := resourceAccessPoint()
			d := r.Data(nil)
			if id, err := AccessPointCreateResourceID(aws.ToString(v.AccessPointArn)); err != nil {
				sweeperErrs = multierror.Append(sweeperErrs, err)
				continue
			} else {
				d.SetId(id)
			}

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error sweeping S3 Access Points (%s): %w", region, err))
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepMultiRegionAccessPoints(region string) error {
	ctx := sweep.Context(region)
	if region != names.USWest2RegionID {
		log.Printf("[WARN] Skipping S3 Multi-Region Access Point sweep for region: %s", region)
		return nil
	}
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.S3ControlClient(ctx)
	accountID := client.AccountID
	input := &s3control.ListMultiRegionAccessPointsInput{
		AccountId: aws.String(accountID),
	}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := s3control.NewListMultiRegionAccessPointsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping S3 Multi-Region Access Point sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing S3 Multi-Region Access Points (%s): %w", region, err)
		}

		for _, v := range page.AccessPoints {
			r := resourceMultiRegionAccessPoint()
			d := r.Data(nil)
			d.SetId(MultiRegionAccessPointCreateResourceID(accountID, aws.ToString(v.Name)))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping S3 Multi-Region Access Points (%s): %w", region, err)
	}

	return nil
}

func sweepObjectLambdaAccessPoints(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.S3ControlClient(ctx)
	accountID := client.AccountID
	input := &s3control.ListAccessPointsForObjectLambdaInput{
		AccountId: aws.String(accountID),
	}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := s3control.NewListAccessPointsForObjectLambdaPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping S3 Object Lambda Access Point sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing S3 Object Lambda Access Points (%s): %w", region, err)
		}

		for _, v := range page.ObjectLambdaAccessPointList {
			r := resourceObjectLambdaAccessPoint()
			d := r.Data(nil)
			d.SetId(ObjectLambdaAccessPointCreateResourceID(accountID, aws.ToString(v.Name)))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping S3 Object Lambda Access Points (%s): %w", region, err)
	}

	return nil
}

func sweepStorageLensConfigurations(region string) error {
	ctx := sweep.Context(region)
	if region == names.USGovEast1RegionID || region == names.USGovWest1RegionID {
		log.Printf("[WARN] Skipping S3 Storage Lens Configuration sweep for region: %s", region)
		return nil
	}
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.S3ControlClient(ctx)
	accountID := client.AccountID
	input := &s3control.ListStorageLensConfigurationsInput{
		AccountId: aws.String(accountID),
	}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := s3control.NewListStorageLensConfigurationsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping S3 Storage Lens Configuration sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing S3 Storage Lens Configurations (%s): %w", region, err)
		}

		for _, v := range page.StorageLensConfigurationList {
			configID := aws.ToString(v.Id)

			if configID == "default-account-dashboard" {
				continue
			}

			r := resourceStorageLensConfiguration()
			d := r.Data(nil)
			d.SetId(StorageLensConfigurationCreateResourceID(accountID, configID))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping S3 Storage Lens Configurations (%s): %w", region, err)
	}

	return nil
}

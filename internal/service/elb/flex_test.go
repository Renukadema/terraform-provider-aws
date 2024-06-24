// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elb_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfelb "github.com/hashicorp/terraform-provider-aws/internal/service/elb"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestExpandListeners(t *testing.T) {
	t.Parallel()

	expanded := []interface{}{
		map[string]interface{}{
			"instance_port":     8000,
			"lb_port":           80,
			"instance_protocol": "http",
			"lb_protocol":       "http",
		},
		map[string]interface{}{
			"instance_port":      8000,
			"lb_port":            80,
			"instance_protocol":  "https",
			"lb_protocol":        "https",
			"ssl_certificate_id": "something",
		},
	}
	listeners, err := tfelb.ExpandListeners(expanded)
	if err != nil {
		t.Fatalf("bad: %#v", err)
	}

	expected := &elb.Listener{
		InstancePort:     aws.Int64(8000),
		LoadBalancerPort: aws.Int64(80),
		InstanceProtocol: aws.String("http"),
		Protocol:         aws.String("http"),
	}

	if !reflect.DeepEqual(listeners[0], expected) {
		t.Fatalf(
			"Got:\n\n%#v\n\nExpected:\n\n%#v\n",
			listeners[0],
			expected)
	}
}

// this test should produce an error from expandlisteners on an invalid
// combination
func TestExpandListeners_invalid(t *testing.T) {
	t.Parallel()

	expanded := []interface{}{
		map[string]interface{}{
			"instance_port":      8000,
			"lb_port":            80,
			"instance_protocol":  "http",
			"lb_protocol":        "http",
			"ssl_certificate_id": "something",
		},
	}
	_, err := tfelb.ExpandListeners(expanded)
	if err != nil {
		// Check the error we got
		if !strings.Contains(err.Error(), `"ssl_certificate_id" may be set only when "protocol"`) {
			t.Fatalf("Got error in TestExpandListeners_invalid, but not what we expected: %s", err)
		}
	}

	if err == nil {
		t.Fatalf("Expected TestExpandListeners_invalid to fail, but passed")
	}
}

func TestFlattenHealthCheck(t *testing.T) {
	t.Parallel()

	cases := []struct {
		Input  *elb.HealthCheck
		Output []map[string]interface{}
	}{
		{
			Input: &elb.HealthCheck{
				UnhealthyThreshold: aws.Int64(10),
				HealthyThreshold:   aws.Int64(10),
				Target:             aws.String("HTTP:80/"),
				Timeout:            aws.Int64(30),
				Interval:           aws.Int64(30),
			},
			Output: []map[string]interface{}{
				{
					"unhealthy_threshold": int64(10),
					"healthy_threshold":   int64(10),
					names.AttrTarget:      "HTTP:80/",
					names.AttrTimeout:     int64(30),
					names.AttrInterval:    int64(30),
				},
			},
		},
	}

	for _, tc := range cases {
		output := tfelb.FlattenHealthCheck(tc.Input)
		if !reflect.DeepEqual(output, tc.Output) {
			t.Fatalf("Got:\n\n%#v\n\nExpected:\n\n%#v", output, tc.Output)
		}
	}
}

func TestExpandInstanceString(t *testing.T) {
	t.Parallel()

	expected := []*elb.Instance{
		{InstanceId: aws.String("test-one")},
		{InstanceId: aws.String("test-two")},
	}

	ids := []interface{}{
		"test-one",
		"test-two",
	}

	expanded := tfelb.ExpandInstanceString(ids)

	if !reflect.DeepEqual(expanded, expected) {
		t.Fatalf("Expand Instance String output did not match.\nGot:\n%#v\n\nexpected:\n%#v", expanded, expected)
	}
}

func TestExpandPolicyAttributes(t *testing.T) {
	t.Parallel()

	expanded := []interface{}{
		map[string]interface{}{
			names.AttrName:  "Protocol-TLSv1",
			names.AttrValue: acctest.CtFalse,
		},
		map[string]interface{}{
			names.AttrName:  "Protocol-TLSv1.1",
			names.AttrValue: acctest.CtFalse,
		},
		map[string]interface{}{
			names.AttrName:  "Protocol-TLSv1.2",
			names.AttrValue: acctest.CtTrue,
		},
	}
	attributes := tfelb.ExpandPolicyAttributes(expanded)

	if len(attributes) != 3 {
		t.Fatalf("expected number of attributes to be 3, but got %d", len(attributes))
	}

	expected := &elb.PolicyAttribute{
		AttributeName:  aws.String("Protocol-TLSv1.2"),
		AttributeValue: aws.String(acctest.CtTrue),
	}

	if !reflect.DeepEqual(attributes[2], expected) {
		t.Fatalf(
			"Got:\n\n%#v\n\nExpected:\n\n%#v\n",
			attributes[2],
			expected)
	}
}

func TestExpandPolicyAttributes_empty(t *testing.T) {
	t.Parallel()

	var expanded []interface{}

	attributes := tfelb.ExpandPolicyAttributes(expanded)

	if len(attributes) != 0 {
		t.Fatalf("expected number of attributes to be 0, but got %d", len(attributes))
	}
}

func TestExpandPolicyAttributes_invalid(t *testing.T) {
	t.Parallel()

	expanded := []interface{}{
		map[string]interface{}{
			names.AttrName:  "Protocol-TLSv1.2",
			names.AttrValue: acctest.CtTrue,
		},
	}
	attributes := tfelb.ExpandPolicyAttributes(expanded)

	expected := &elb.PolicyAttribute{
		AttributeName:  aws.String("Protocol-TLSv1.2"),
		AttributeValue: aws.String(acctest.CtFalse),
	}

	if reflect.DeepEqual(attributes[0], expected) {
		t.Fatalf(
			"Got:\n\n%#v\n\nExpected:\n\n%#v\n",
			attributes[0],
			expected)
	}
}

func TestFlattenPolicyAttributes(t *testing.T) {
	t.Parallel()

	cases := []struct {
		Input  []*elb.PolicyAttributeDescription
		Output []interface{}
	}{
		{
			Input: []*elb.PolicyAttributeDescription{
				{
					AttributeName:  aws.String("Protocol-TLSv1.2"),
					AttributeValue: aws.String(acctest.CtTrue),
				},
			},
			Output: []interface{}{
				map[string]string{
					names.AttrName:  "Protocol-TLSv1.2",
					names.AttrValue: acctest.CtTrue,
				},
			},
		},
	}

	for _, tc := range cases {
		output := tfelb.FlattenPolicyAttributes(tc.Input)
		if !reflect.DeepEqual(output, tc.Output) {
			t.Fatalf("Got:\n\n%#v\n\nExpected:\n\n%#v", output, tc.Output)
		}
	}
}

package main

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

// To describe the status of an instance
//
// This example describes the current status of the specified instance.

func ExampleEC2_DescribeInstanceStatus_shared00() {
	svc := ec2.New(session.New())
	input := &ec2.DescribeInstanceStatusInput{
		InstanceIds: []*string{
			aws.String("i-1234567890abcdef0"),
		},
	}

	result, err := svc.DescribeInstanceStatus(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			fmt.Println(err.Error())
		}
		return
	}

	fmt.Println(result)
}

// To describe an Amazon EC2 instance
//
// This example describes the specified instance.
func ExampleEC2_DescribeInstances_shared00() {
	svc := ec2.New(session.New())
	input := &ec2.DescribeInstancesInput{
		InstanceIds: []*string{
			aws.String("i-1234567890abcdef0"),
		},
	}

	result, err := svc.DescribeInstances(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			fmt.Println(err.Error())
		}
		return
	}

	fmt.Println(result)
}

// To describe the instances with a specific tag
//
// This example describes the instances with the Purpose=test tag.
func ExampleEC2_DescribeInstances_shared02() {
	svc := ec2.New(session.New())
	input := &ec2.DescribeInstancesInput{
		Filters: []*ec2.Filter{
			{
				Name: aws.String("tag:syncbak-office-hours"),
				Values: []*string{
					aws.String("test"),
				},
			},
		},
	}

	result, err := svc.DescribeInstances(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			fmt.Println(err.Error())
		}
		return
	}

	fmt.Println(result)
}

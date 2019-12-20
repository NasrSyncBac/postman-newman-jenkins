package main

import (
	"fmt"
	"log"
	"os"

	"text/template"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

var tpl *template.Template

func init() {
	tpl = template.Must(template.ParseFiles("tpl.gohtml"))
}
func main() {
	// describeStoppedEC2("i-001baaae3654b441f")
	describeAutoScalingGroup()
	//describeTaggedInstance()
}

type data struct {
	ID    string
	State string
}

func describeAutoScalingGroup() {

	nf, err := os.Create("index.html")

	if err != nil {
		log.Fatalln("error creating file", err)
	}
	svc := ec2.New(session.New())

	result, err := svc.DescribeInstances(nil)
	if err != nil {
		fmt.Println("Error", err)
	} else {
		// fmt.Println(result)
		for _, res := range result.Reservations {
			for _, instance := range res.Instances {

				if *instance.State.Name == "stopped" || *instance.State.Name == "terminated" {

					state := *instance.State.Name
					id := *instance.InstanceId

					// fmt.Println(id)
					// fmt.Println(state)
					results := data{
						ID:    id,
						State: state,
					}

					defer nf.Close()

					err = tpl.Execute(nf, results)

					if err != nil {
						log.Fatalln(err)
					}
					// fmt.Println(*instance.State.Name)
					// fmt.Println(*instance.InstanceId)

				}
			}

		}
	}

}

// func describeEC2(instance string) {

// 	svc := ec2.New(session.New())
// 	input := &ec2.DescribeInstanceStatusInput{
// 		InstanceIds: []*string{
// 			aws.String(instance),
// 		},
// 	}

// 	result, err := svc.DescribeInstanceStatus(input)
// 	if err != nil {
// 		if aerr, ok := err.(awserr.Error); ok {
// 			switch aerr.Code() {
// 			default:
// 				fmt.Println(aerr.Error())
// 			}
// 		} else {
// 			fmt.Println(err.Error())
// 		}
// 		return
// 	}
// 	fmt.Println(result)
// }

// func describeStoppedEC2(instance string) {

// 	svc := ec2.New(session.New())
// 	input := &ec2.DescribeInstancesInput{
// 		InstanceIds: []*string{
// 			aws.String(instance),
// 		},
// 		Filters: []*ec2.Filter{
// 			&ec2.Filter{
// 				Name: aws.String("instance-state-name"),
// 				Values: []*string{
// 					aws.String("running"),
// 					aws.String("pending"),
// 					aws.String("stopped"),
// 					aws.String("terminated"),
// 					aws.String("shutting down"),
// 					aws.String("stopping"),
// 				},
// 			},
// 		},
// 	}
// 	result, err := svc.DescribeInstances(input)
// 	if err != nil {
// 		if aerr, ok := err.(awserr.Error); ok {
// 			switch aerr.Code() {
// 			default:
// 				fmt.Println(aerr.Error())
// 			}
// 		} else {
// 			fmt.Println(err.Error())
// 		}
// 		return
// 	}

// 	//fmt.Println(result)

// 	for _, res := range result.Reservations {
// 		for _, instance := range res.Instances {
// 			fmt.Println(*instance.InstanceId)

// 			if *instance.State.Name == "terminated" {
// 				fmt.Println(*instance.State.Name)
// 			}

// 		}
// 	}
// }

// func check(e error) {

// 	if e != nil {
// 		panic(e)
// 	}
// }

// func describeTaggedInstance() {
// 	svc := ec2.New(session.New())
// 	input := &ec2.DescribeInstancesInput{
// 		Filters: []*ec2.Filter{
// 			{
// 				Name: aws.String("tag:syncbak-office-hours"),
// 				Values: []*string{
// 					aws.String("test"),
// 				},
// 			},
// 		},
// 	}

// 	result, err := svc.DescribeInstances(input)
// 	if err != nil {
// 		if aerr, ok := err.(awserr.Error); ok {
// 			switch aerr.Code() {
// 			default:
// 				fmt.Println(aerr.Error())
// 			}
// 		} else {
// 			// Print the error, cast err to awserr.Error to get the Code and
// 			// Message from an error.
// 			fmt.Println(err.Error())
// 		}
// 		return
// 	}

// 	fmt.Println(result)
// }

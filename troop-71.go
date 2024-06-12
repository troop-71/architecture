package main

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsec2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsecs"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsecspatterns"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsrds"
	// "github.com/aws/aws-cdk-go/awscdk/v2/awssqs"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

type Troop71StackProps struct {
	awscdk.StackProps
}

func NewTroop71Stack(scope constructs.Construct, id string, props *Troop71StackProps) awscdk.Stack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	vpc := awsec2.NewVpc(stack, jsii.String("vpc"), &awsec2.VpcProps{
		SubnetConfiguration: &[]*awsec2.SubnetConfiguration{{
			SubnetType: awsec2.SubnetType_PUBLIC,
			Name:       jsii.String("subnet"),
		}},
	})

	postgres := awsrds.NewDatabaseInstance(stack, jsii.String("rds"), &awsrds.DatabaseInstanceProps{
		Vpc:          vpc,
		InstanceType: awsec2.InstanceType_Of(awsec2.InstanceClass_T4G, awsec2.InstanceSize_MICRO),
		Engine:       awsrds.DatabaseInstanceEngine_POSTGRES(),
		VpcSubnets: &awsec2.SubnetSelection{
			SubnetType: awsec2.SubnetType_PUBLIC,
		},
		DatabaseName: jsii.String("wiki"),
	})

	cluster := awsecs.NewCluster(stack, jsii.String("cluster"), &awsecs.ClusterProps{
		Vpc: vpc,
	})

	//cluster.Connections().AllowToAnyIpv4(awsec2.Port_HTTPS(), jsii.String("allow https"))

	ecs := awsecspatterns.NewApplicationLoadBalancedFargateService(stack, jsii.String("wikijs"), &awsecspatterns.ApplicationLoadBalancedFargateServiceProps{
		Cluster:        cluster,
		AssignPublicIp: jsii.Bool(true),
		//CircuitBreaker: &awsecs.DeploymentCircuitBreaker{
		//	Enable: jsii.Bool(false),
		//},
		DesiredCount:         jsii.Number(0),
		EnableECSManagedTags: jsii.Bool(true),
		TaskImageOptions: &awsecspatterns.ApplicationLoadBalancedTaskImageOptions{
			Image: awsecs.ContainerImage_FromRegistry(
				jsii.String("ghcr.io/requarks/wiki:2"),
				&awsecs.RepositoryImageProps{},
			),
			Environment: &map[string]*string{
				//"DB_TYPE": jsii.String("postgres"),
			},
			Secrets: &map[string]awsecs.Secret{
				"DB_PASS": awsecs.Secret_FromSecretsManager(postgres.Secret(), jsii.String("password")),
				"DB_USER": awsecs.Secret_FromSecretsManager(postgres.Secret(), jsii.String("username")),
				"DB_PORT": awsecs.Secret_FromSecretsManager(postgres.Secret(), jsii.String("port")),
				"DB_HOST": awsecs.Secret_FromSecretsManager(postgres.Secret(), jsii.String("host")),
				"DB_TYPE": awsecs.Secret_FromSecretsManager(postgres.Secret(), jsii.String("engine")),
			},
		},
	})
	postgres.Connections().AllowDefaultPortFrom(
		ecs.Service(),
		//awsec2.Port_TcpRange(jsii.Number(5432), jsii.Number(5432)),
		jsii.String("allow cluster to rds"),
	)

	//importedHostedZone := awsroute53.HostedZone_FromHostedZoneAttributes(
	//	stack,
	//	jsii.String("hosted zone"),
	//	&awsroute53.HostedZoneAttributes{
	//		HostedZoneId: jsii.String("Z02543953LEPR7NK5UEHN"),
	//		ZoneName:     jsii.String("troop-71.com"),
	//	},
	//)

	//awsroute53.NewARecord(stack, jsii.String("A record"), &awsroute53.ARecordProps{
	//	Zone:   importedHostedZone,
	//})

	return stack
}

func main() {
	defer jsii.Close()

	app := awscdk.NewApp(nil)

	NewTroop71Stack(app, "Troop71Stack", &Troop71StackProps{
		awscdk.StackProps{
			Env: env(),
		},
	})

	app.Synth(nil)
}

// env determines the AWS environment (account+region) in which our stack is to
// be deployed. For more information see: https://docs.aws.amazon.com/cdk/latest/guide/environments.html
func env() *awscdk.Environment {
	// If unspecified, this stack will be "environment-agnostic".
	// Account/Region-dependent features and context lookups will not work, but a
	// single synthesized template can be deployed anywhere.
	//---------------------------------------------------------------------------
	return nil

	// Uncomment if you know exactly what account and region you want to deploy
	// the stack to. This is the recommendation for production stacks.
	//---------------------------------------------------------------------------
	// return &awscdk.Environment{
	//  Account: jsii.String("123456789012"),
	//  Region:  jsii.String("us-east-1"),
	// }

	// Uncomment to specialize this stack for the AWS Account and Region that are
	// implied by the current CLI configuration. This is recommended for dev
	// stacks.
	//---------------------------------------------------------------------------
	// return &awscdk.Environment{
	//  Account: jsii.String(os.Getenv("CDK_DEFAULT_ACCOUNT")),
	//  Region:  jsii.String(os.Getenv("CDK_DEFAULT_REGION")),
	// }
}

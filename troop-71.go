package main

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsec2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsecs"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsecspatterns"
	"github.com/aws/aws-cdk-go/awscdk/v2/awselasticloadbalancingv2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsrds"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsroute53"
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
			Name:       jsii.String("public subnet"),
			SubnetType: awsec2.SubnetType_PUBLIC,
		}},
		NatGateways: jsii.Number(0),
		MaxAzs:      jsii.Number(1),
		IpAddresses: awsec2.IpAddresses_Cidr(jsii.String("10.1.0.0/16")),
	})

	engine := awsrds.DatabaseInstanceEngine_Postgres(&awsrds.PostgresInstanceEngineProps{
		Version: awsrds.PostgresEngineVersion_VER_16_3(),
	})

	postgres := awsrds.NewDatabaseInstance(stack, jsii.String("rds"), &awsrds.DatabaseInstanceProps{
		Vpc:          vpc,
		InstanceType: awsec2.InstanceType_Of(awsec2.InstanceClass_T4G, awsec2.InstanceSize_MICRO),
		Engine:       engine,
		DatabaseName: jsii.String("wiki"),
		Parameters: &map[string]*string{
			"rds.force_ssl": jsii.String("0"),
		},
		VpcSubnets: &awsec2.SubnetSelection{
			SubnetType: awsec2.SubnetType_PUBLIC,
		},
		BackupRetention:  awscdk.Duration_Days(jsii.Number(7)),
		AllocatedStorage: jsii.Number(20),
	})

	importedHostedZone := awsroute53.HostedZone_FromHostedZoneAttributes(
		stack,
		jsii.String("hosted zone"),
		&awsroute53.HostedZoneAttributes{
			HostedZoneId: jsii.String("Z02543953LEPR7NK5UEHN"),
			ZoneName:     jsii.String("troop-71.com"),
		},
	)

	ecs := awsecspatterns.NewApplicationLoadBalancedFargateService(stack, jsii.String("wikijs"), &awsecspatterns.ApplicationLoadBalancedFargateServiceProps{
		AssignPublicIp:       jsii.Bool(true),
		EnableECSManagedTags: jsii.Bool(true),
		RedirectHTTP:         jsii.Bool(true),
		Protocol:             awselasticloadbalancingv2.ApplicationProtocol_HTTPS,
		Cpu:                  jsii.Number(512),
		MemoryLimitMiB:       jsii.Number(1024),
		CapacityProviderStrategies: &[]*awsecs.CapacityProviderStrategy{{
			CapacityProvider: jsii.String("FARGATE_SPOT"),
			Weight:           jsii.Number(1),
		}},

		HealthCheck: &awsecs.HealthCheck{
			Command: &[]*string{
				jsii.String("CMD-SHELL"),
				jsii.String("curl -f http://localhost:3000/healthz || exit 1"),
			}},
		TaskImageOptions: &awsecspatterns.ApplicationLoadBalancedTaskImageOptions{
			ContainerPort: jsii.Number(3000),
			Image: awsecs.ContainerImage_FromRegistry(
				jsii.String("ghcr.io/requarks/wiki:2"),
				&awsecs.RepositoryImageProps{},
			),
			Environment: &map[string]*string{
				"HA":      jsii.String("true"),
				"DB_NAME": jsii.String("wiki"),
			},
			Secrets: &map[string]awsecs.Secret{
				"DB_PASS": awsecs.Secret_FromSecretsManager(postgres.Secret(), jsii.String("password")),
				"DB_USER": awsecs.Secret_FromSecretsManager(postgres.Secret(), jsii.String("username")),
				"DB_PORT": awsecs.Secret_FromSecretsManager(postgres.Secret(), jsii.String("port")),
				"DB_HOST": awsecs.Secret_FromSecretsManager(postgres.Secret(), jsii.String("host")),
				"DB_TYPE": awsecs.Secret_FromSecretsManager(postgres.Secret(), jsii.String("engine")),
			},
		},
		Vpc:          vpc,
		ListenerPort: jsii.Number(443),
		DomainName:   jsii.String("troop-71.com"),
		DomainZone:   importedHostedZone,
	})

	postgres.Connections().AllowDefaultPortFrom(
		ecs.Service(),
		jsii.String("allow cluster to rds"),
	)

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

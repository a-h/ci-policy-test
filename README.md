# ci-policy-test

This repo is being used to try out some ideas around CI/CD pipelines for Serverless applications.

One of the challenges of deploying Serverless applications within larger organisations is that Serverless applications blur the boundaries between application and infrastructure. Releasing a minor update to a Serverless application will regularly include changes to infrastructure along with changes to application behaviour.

In a typical container-based application, a relatively static network and Kubernetes configuration is setup at the infrastructure level, and application code runs inside the cluster, but in a Serverless application, even just adding a new HTTP endpoint may involve adding a new route to API Gateway, connecting that to a new Lambda function, perhaps creating a new DynamoDB table and, importantly, creating an IAM role that allows the function to access that table.

Tools like Serverless Framework and SAM automatically deploy the IAM role for you, when you use `sls deploy` etc. However, this relies on the user running the deployment to have permissions to create the new infrastructure and IAM roles. Many organisations have completed automated deployments across development, testing and production environments, typically running in separate AWS accounts to make it easier to control and audit access.

Most organisations create a CI/CD pipeline to deploy changes to automate deployments, and also prevent needing to give developers write access to infrastructure. In less mature organisations, the CI/CD user is basically an Administrator level user, with permission to do just about anything.

If the developers don't have write access to infrastructure, but can make changes to the source code, then they could commit code into the repository that changes their permissions, and give themselves admin rights. To mitigate against this (and other problems), an approval step is usually carried out. To prevent wasting the time of reviewers, reviews generally don't happen until unit and integration tests have executed in the CI/CD pipeline. CI/CD pipelines often deploy test environments in AWS to run integration tests, so basically, the code review can happen once the damage has been done - another great reason to have a separated dev and production AWS account. For a developer, a late night source code push, followed by destroying the branch (or doing a force push to rewrite history) may be enough to cover up tracks.

To prevent the "evil developer" scenario, organisations pick from several options:

1. Make sure that the CI/CD pipeline only runs in the test environment until a review has been completed.

This is attractive because it's simple. It avoids having to do anything complicated with IAM permissions. But... it might not be secure enough to meet your security or compliance requirements.

2. Rely on automated checking for unusual IAM changes using AWS Config and GuardDuty.

Makes sense, but it could be too slow if human action is required to resolve. I've seen attacks start within 10 minutes of AWS credentials being leaked.

3. Don't give the CI/CD user permission to update or assign IAM roles, and route permission changes through a dedicated security team.

There are typically fewer people in the security team than there are developers, meaning that overall, fewer people have access here. However, having to explain every minor change can be very frustrating for teams that are attempting to move at velocity. The security teams would also need to be experts in the uses and abuses of AWS to avoid approving something that looked innocuous, but was actually an attack.

4. Agree with the security team the maximum set of permissions that would be sensible for a team, and then assign an IAM Permission Boundary to the CI/CD user to configure the maximum set of permissions the CI/CD user can have. Next, give the development team permission to change that CI/CD user's permissions as much as they like within that permission boundary - only changes that fall within the permission boundary will take effect. This encourages the development team to use a minimal role for their CI/CD user, because they can change it whenever they need to.

The use of a Permission Boundary seems to be the sweet spot, but it comes with a lot of complexity. The Permission Boundary needs to be owned by a security team, platform team, or a subset of the development team that understand the use of Permission Boundaries, and have deep understanding of IAM. The Permission Boundary also needs to be written carefully to prevent the possibility of privilege escalation attacks.

This repo contains several things I'm working on in this area:

* A "starter" permission boundary CloudFormation template
  * ./serverless/serverless-permission-boundary.yaml
* A "starter" CI user CloudFormation template
  * ./serverless/squad-ci-user.yaml
* A "hello world" Serverless app for testing that uses the Permission Boundary, and the custom deployment bucket
  * ./serverless/app

It's really difficult to review IAM configuration because the intent of the configuration is not readily visible. To try and make the intent of permissions clear to auditors, I started writing a program that could be executed in a CI/CD pipeline to validate that privilege escalation wasn't possible, and that the roles were correctly configured for the use case.

```go
p := NewPrincipal(iamSvc, callerIdentityARN)
p.CanCreateRoleWithBoundary()
p.CannotCreateUsers()
p.CannotLaunchEC2Instances()
p.CanLaunchStacksInIreland()
p.CanLaunchStacksInNorthVirginia()
p.CannotLaunchStacksOutsideIrelandAndNorthVirginia()
p.CanListS3Buckets()
p.CannotReadDynamoDBData()
p.CannotModifyDynamoDBData()
p.CannotExecuteDynamoDBTransactions()
p.CanPassRoleToLambda()
p.CanDeleteRole()
```

A test looks like this. It's just IAM policy simulator really, but the test also includes a reason "why" it's important to make it easier for people to understand the check.

```go
func (p *Principal) CannotLaunchEC2Instances() {
	testName := "CannotLaunchEC2Instances"
	input := &iam.SimulatePrincipalPolicyInput{
		PolicySourceArn: p.callerIdentityARN,
		ActionNames:     aws.StringSlice([]string{"ec2:LaunchInstance"}),
	}
	expected := Expected{
		Allowed: false,
	}
	reason := `It's unusual for a Serverless application to launch EC2 instances.`
	p.tests = append(p.tests, NewTest(testName, input, expected, reason))
}
```

I chose Go to write it in so that it can be built to run without dependenices, to make it easier to drop into a CI/CD pipeline, or run on any box (Linux, Mac, Windows).

## Github Action

This action uses AWS credentials supplied from the Github organisation or repository level to run the tests in a CI pipeline.

### Inputs

#### `AWS_ACCESS_KEY_ID`

**Required** AWS_ACCESS_KEY_ID. Reference the variable in your repo or organisation, whatever you've called it. Default `''`.

#### `AWS_SECRET_ACCESS_KEY`

**Required** AWS_SECRET_ACCESS_KEY. Reference the secret, whatever you've called it. Default `''`.

### Example usage

```yaml
jobs:
  ci-policy-test:
    name: CI Policy Test
    runs-on: ubuntu-latest
    steps:
      - uses: a-h/ci-policy-test@v1
        with:
          AWS_ACCESS_KEY_ID: ${{ secrets.AWS_DEV_ACCESS_KEY_ID }}
          AWS_SECRET_ACCESS_KEY: ${{ secrets.AWS_DEV_ACCESS_SECRET }}
```
## Next steps

I've made the permission boundary opinionated, in that each squad must create resources that have their squad prefix in it. The permission boundary could be made to be a squad level one, but, I think it might be overkill.

I plan to create a CDK example to test out whether the boundaries also work for CDK deployments.

Write up a blog post when I've got these exact templates running in production (I have similar ones running in prod, but I want to be careful) and ironed out any wrinkles.

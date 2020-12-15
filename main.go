package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/sts"
)

func main() {
	sess := session.New()

	// aws sts get-caller-identity
	stsSvc := sts.New(sess)
	callerIdentity, err := stsSvc.GetCallerIdentity(&sts.GetCallerIdentityInput{})
	if err != nil {
		fmt.Println("failed call to GetCallerIdentity:", err)
		os.Exit(1)
	}
	callerIdentityARN, err := arn.Parse(*callerIdentity.Arn)
	if err != nil {
		fmt.Println("failed to parse GetCallerIdentity ARN:", err)
		os.Exit(1)
	}

	// aws iam list-roles
	// If the current caller identity is not a user, but is an assumed role, we need to find out the policies
	// attached to the role to be able to simulate the policies.
	iamSvc := iam.New(sess)
	if callerIdentityARN.Service == "iam" {
		fmt.Printf("The current principal is an IAM entity: %q\n", callerIdentityARN.String())
	}
	if callerIdentityARN.Service == "sts" &&
		strings.HasPrefix(callerIdentityARN.Resource, "assumed-role/") {
		fmt.Printf("The current principal is an assumed role, locating IAM entity: %q\n", callerIdentityARN.String())
		roleName := strings.SplitN(callerIdentityARN.Resource, "/", 3)[1]
		role, err := iamSvc.GetRole(&iam.GetRoleInput{RoleName: &roleName})
		if err != nil {
			fmt.Println("failed call to GetRoleInput:", err)
			os.Exit(1)
		}
		callerIdentityARN, err = arn.Parse(*role.Role.Arn)
		if err != nil {
			fmt.Println("failed to parse the role ARN returned by GetRole:", err)
			os.Exit(1)
		}
		fmt.Printf("Located IAM entity: %q\n", callerIdentityARN.String())
	}

	// aws iam simulate-principal-policy --action-names dynamodb:CreateBackup --policy-source-arn arn:aws:iam:::role/Administrator
	// Now it's possible to test the actions that the entity can carry out.
	p := NewPrincipal(iamSvc, callerIdentityARN)
	p.CannotCreateUsers()
	p.CannotLaunchEC2Instances()
	p.CanCreateNewPolicies()
	success := p.Test()
	if !success {
		os.Exit(1)
	}
}

func NewPrincipal(svc *iam.IAM, callerIdentityARN arn.ARN) *Principal {
	return &Principal{
		svc:               svc,
		callerIdentityARN: aws.String(callerIdentityARN.String()),
		tests:             make([]*Test, 0),
	}
}

type Principal struct {
	svc               *iam.IAM
	callerIdentityARN *string
	tests             []*Test
}

func NewTest(name string, input *iam.SimulatePrincipalPolicyInput, e Expected, reason string) *Test {
	return &Test{
		Name:     name,
		Input:    input,
		Expected: e,
		Reason:   reason,
	}
}

type Test struct {
	Name     string
	Input    *iam.SimulatePrincipalPolicyInput
	Expected Expected
	Reason   string
}

type Expected struct {
	Decision string
}

func (p *Principal) runTest(t *Test) (passed bool, results []*iam.EvaluationResult, err error) {
	passed = true
	pager := func(spr *iam.SimulatePolicyResponse, lastPage bool) (carryOn bool) {
		for i := 0; i < len(spr.EvaluationResults); i++ {
			if t.Expected.Decision != *spr.EvaluationResults[i].EvalDecision {
				passed = false
				results = append(results, spr.EvaluationResults[i])
				return false
			}
		}
		return true
	}
	err = p.svc.SimulatePrincipalPolicyPages(t.Input, pager)
	return
}

func (p *Principal) CannotCreateUsers() {
	testName := "CannotCreateUsers"
	input := &iam.SimulatePrincipalPolicyInput{
		PolicySourceArn: p.callerIdentityARN,
		ActionNames:     aws.StringSlice([]string{"iam:CreateUser"}),
	}
	expected := Expected{
		Decision: "denied",
	}
	reason := `A CI pipeline should not be able to create new users.`
	p.tests = append(p.tests, NewTest(testName, input, expected, reason))
}

func (p *Principal) CannotLaunchEC2Instances() {
	testName := "CannotLaunchEC2Instances"
	input := &iam.SimulatePrincipalPolicyInput{
		PolicySourceArn: p.callerIdentityARN,
		ActionNames:     aws.StringSlice([]string{"ec2:LaunchInstance"}),
	}
	expected := Expected{
		Decision: "denied",
	}
	reason := `It's unusual for a Serverless application to launch EC2 instances.`
	p.tests = append(p.tests, NewTest(testName, input, expected, reason))
}

func (p *Principal) CanCreateNewPolicies() {
	testName := "CanCreateNewPolicies"
	input := &iam.SimulatePrincipalPolicyInput{
		PolicySourceArn: p.callerIdentityARN,
		ActionNames: aws.StringSlice([]string{
			"iam:CreatePolicy",
		}),
	}
	expected := Expected{
		Decision: "allowed",
	}
	reason := `It's relatively risk free to create a new policy, because it doesn't take action if it's not attached to a resource.`
	p.tests = append(p.tests, NewTest(testName, input, expected, reason))
}

func renderResult(er []*iam.EvaluationResult) string {
	var sb strings.Builder
	for i := 0; i < len(er); i++ {
		bytes, _ := json.Marshal(er[i])
		m := make(map[string]interface{})
		json.Unmarshal(bytes, &m)
		for k, v := range m {
			if v == nil {
				delete(m, k)
			}
		}
		bytes, _ = json.MarshalIndent(m, "     ", " ")
		sb.Write(bytes)
		sb.WriteRune('\n')
	}
	return sb.String()
}

const tick = string(rune(0x2714))
const cross = string(rune(0x2717))

func (p *Principal) Test() (success bool) {
	success = true
	for _, t := range p.tests {
		t := t
		passed, result, err := p.runTest(t)
		if err != nil {
			fmt.Printf(" %s %s ERROR %v\n", cross, t.Name, err)
			success = false
			continue
		}
		if !passed {
			fmt.Printf(" %s %s FAIL\n     %v", cross, t.Name, renderResult(result))
			success = false
			continue
		}
		fmt.Printf(" %s %s PASS\n", tick, t.Name)
	}
	return
}

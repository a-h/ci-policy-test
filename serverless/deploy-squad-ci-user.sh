# Deploy the template for a squad.
aws cloudformation deploy \
--stack-name=squad-ci-user-mysquad \
--template-file=./squad-ci-user.yaml \
--capabilities=CAPABILITY_NAMED_IAM \
--region=eu-west-1 \
--parameter-overrides Squad=MySquad

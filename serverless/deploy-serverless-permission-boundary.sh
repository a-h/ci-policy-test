# Deploy the global permission boundary.
aws cloudformation deploy \
--stack-name=serverless-permission-boundary \
--template-file=serverless-permission-boundary.yaml  \
--capabilities=CAPABILITY_NAMED_IAM  \
--region=eu-west-1 \
--parameter-overrides Squad=MySquad

# CI Policy Test docker action

This action uses AWS credentials supplied from the Github organisation or repository level to run the tests in a CI pipeline.

## Inputs

### `AWS_ACCESS_KEY_ID`

**Required** AWS_ACCESS_KEY_ID. Reference the variable in your repo or organisation, whatever you've called it. Default `''`.

### `AWS_SECRET_ACCESS_KEY`

**Required** AWS_SECRET_ACCESS_KEY. Reference the secret, whatever you've called it. Default `''`.

## Example usage

uses: actions/ci-policy-test-action@v1
with:
  AWS_ACCESS_KEY_ID: ${{ secrets.AWS_DEV_CI_USER_ACCESS_KEY }}
  AWS_SECRET_ACCESS_KEY: ${{ secrets.AWS_DEV_CI_USER_SECRET }}
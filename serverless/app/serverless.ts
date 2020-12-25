import type { AWS } from '@serverless/typescript';

const serverlessConfiguration: AWS = {
  service: 'ci-policy-test-app',
  frameworkVersion: '2',
  custom: {
    webpack: {
      webpackConfig: './webpack.config.js',
      includeModules: true
    }
  },
  // Add the serverless-webpack plugin
  plugins: ['serverless-webpack'],
  provider: {
    name: 'aws',
    runtime: 'nodejs12.x',
    region: 'eu-west-1',
    deploymentBucket: {
      name: "serverless-ci-serverlessdeploymentbucket-qrrwxuypevdz",
    },
    rolePermissionsBoundary: "arn:aws:iam::180466524585:policy/global-ci-serverless-permission-boundary",
    apiGateway: {
      minimumCompressionSize: 1024,
    },
    environment: {
      AWS_NODEJS_CONNECTION_REUSE_ENABLED: '1',
    },
  },
  functions: {
    hello: {
      handler: 'handler.hello',
      events: [
        {
          http: {
            method: 'get',
            path: 'hello',
          }
        }
      ]
    }
  }
}

module.exports = serverlessConfiguration;

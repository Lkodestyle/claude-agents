# GitLab CI/CD Templates

## Pipeline básico (Node.js)
```yaml
# .gitlab-ci.yml
stages:
  - test
  - build
  - deploy

variables:
  NODE_VERSION: "20"

test:
  stage: test
  image: node:${NODE_VERSION}-alpine
  cache:
    key: ${CI_COMMIT_REF_SLUG}
    paths:
      - node_modules/
  before_script:
    - npm ci
  script:
    - npm run lint
    - npm test
  coverage: '/Lines\s*:\s*(\d+\.?\d*)%/'

build:
  stage: build
  image: docker:24
  services:
    - docker:24-dind
  variables:
    DOCKER_TLS_CERTDIR: "/certs"
  before_script:
    - docker login -u $CI_REGISTRY_USER -p $CI_REGISTRY_PASSWORD $CI_REGISTRY
  script:
    - docker build -t $CI_REGISTRY_IMAGE:$CI_COMMIT_SHA .
    - docker push $CI_REGISTRY_IMAGE:$CI_COMMIT_SHA
  rules:
    - if: $CI_COMMIT_BRANCH == "main"

deploy_staging:
  stage: deploy
  image: alpine:latest
  environment:
    name: staging
    url: https://staging.example.com
  before_script:
    - apk add --no-cache aws-cli
  script:
    - aws ecs update-service --cluster staging --service app --force-new-deployment
  rules:
    - if: $CI_COMMIT_BRANCH == "develop"

deploy_production:
  stage: deploy
  image: alpine:latest
  environment:
    name: production
    url: https://example.com
  before_script:
    - apk add --no-cache aws-cli
  script:
    - aws ecs update-service --cluster production --service app --force-new-deployment
  rules:
    - if: $CI_COMMIT_BRANCH == "main"
  when: manual
```

## Pipeline con Multi-Environment
```yaml
# .gitlab-ci.yml
stages:
  - test
  - build
  - deploy

# Templates
.deploy_template: &deploy_template
  image: alpine:latest
  before_script:
    - apk add --no-cache aws-cli curl
  script:
    - aws ecs update-service --cluster $CLUSTER --service $SERVICE --force-new-deployment
    - |
      for i in $(seq 1 30); do
        STATUS=$(aws ecs describe-services --cluster $CLUSTER --services $SERVICE --query 'services[0].deployments[0].rolloutState' --output text)
        if [ "$STATUS" = "COMPLETED" ]; then
          echo "Deployment successful!"
          exit 0
        fi
        echo "Waiting... ($i/30)"
        sleep 10
      done
      echo "Deployment timeout"
      exit 1

# Jobs
test:
  stage: test
  image: node:20-alpine
  script:
    - npm ci
    - npm test

build:
  stage: build
  image: docker:24
  services:
    - docker:24-dind
  script:
    - docker login -u $CI_REGISTRY_USER -p $CI_REGISTRY_PASSWORD $CI_REGISTRY
    - docker build -t $CI_REGISTRY_IMAGE:$CI_COMMIT_SHA .
    - docker push $CI_REGISTRY_IMAGE:$CI_COMMIT_SHA
  rules:
    - if: $CI_COMMIT_BRANCH =~ /^(main|develop)$/

deploy_dev:
  <<: *deploy_template
  stage: deploy
  variables:
    CLUSTER: dev-cluster
    SERVICE: app-service
  environment:
    name: development
  rules:
    - if: $CI_COMMIT_BRANCH == "develop"

deploy_staging:
  <<: *deploy_template
  stage: deploy
  variables:
    CLUSTER: staging-cluster
    SERVICE: app-service
  environment:
    name: staging
  rules:
    - if: $CI_COMMIT_BRANCH == "main"

deploy_production:
  <<: *deploy_template
  stage: deploy
  variables:
    CLUSTER: prod-cluster
    SERVICE: app-service
  environment:
    name: production
  rules:
    - if: $CI_COMMIT_BRANCH == "main"
  when: manual
```

## Pipeline con Terraform
```yaml
# .gitlab-ci.yml
stages:
  - validate
  - plan
  - apply

image:
  name: hashicorp/terraform:1.6
  entrypoint: [""]

variables:
  TF_ROOT: ${CI_PROJECT_DIR}/infrastructure

before_script:
  - cd ${TF_ROOT}
  - terraform init

validate:
  stage: validate
  script:
    - terraform validate
    - terraform fmt -check

plan:
  stage: plan
  script:
    - terraform plan -out=plan.tfplan
  artifacts:
    paths:
      - ${TF_ROOT}/plan.tfplan
    expire_in: 1 week

apply:
  stage: apply
  script:
    - terraform apply -auto-approve plan.tfplan
  dependencies:
    - plan
  rules:
    - if: $CI_COMMIT_BRANCH == "main"
  when: manual
  environment:
    name: production
```

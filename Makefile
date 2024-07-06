ENV ?= dev
include .env
-include .env.${ENV}

enable_apis:
	gcloud services enable artifactregistry.googleapis.com cloudscheduler.googleapis.com cloudtasks.googleapis.com run.googleapis.com

create_repository:
	gcloud artifacts repositories create ${REPO_NAME} \
	--repository-format=docker \
	--location=${REPO_LOCATION} \
	--immutable-tags

configure_repository_auth:
	gcloud auth configure-docker ${REPO_LOCATION}-docker.pkg.dev --quiet

build_manager:
	cd manager && CGO_ENABLED=0 go build && podman build -t ${REPO_LOCATION}-docker.pkg.dev/${PROJECT_ID}/${REPO_NAME}/manager:${version} .

push_manager:
	podman push ${REPO_LOCATION}-docker.pkg.dev/${PROJECT_ID}/${REPO_NAME}/manager:${version}

update_manager: build_manager push_manager

build_notifier:
	cd notifier && CGO_ENABLED=0 go build && podman build -t ${REPO_LOCATION}-docker.pkg.dev/${PROJECT_ID}/${REPO_NAME}/notifier:${version} .

push_notifier:
	podman push ${REPO_LOCATION}-docker.pkg.dev/${PROJECT_ID}/${REPO_NAME}/notifier:${version}

update_notifier: build_notifier push_notifier

apply_terraform:
	cd terraform && terraform apply
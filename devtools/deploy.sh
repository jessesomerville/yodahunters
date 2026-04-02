#!/bin/bash
set -e

# Configuration
PROJECT_ID="yodahunters"
REGION="us-central1"
SERVICE_NAME="yodahunters"
IMAGE_NAME="${REGION}-docker.pkg.dev/${PROJECT_ID}/${SERVICE_NAME}/${SERVICE_NAME}:latest"
SERVICE_ACCOUNT="yodahunters-runner@${PROJECT_ID}.iam.gserviceaccount.com"
DB_INSTANCE="${PROJECT_ID}:${REGION}:yodahunters-db"
DB_NAME="yodahunters-db"
DB_USER="yodahunters-runner@${PROJECT_ID}.iam" # IAM format for Postgres

echo "Starting deployment for ${SERVICE_NAME}..."

# 1. Build and push image using Cloud Build
echo "Building container image..."
gcloud builds submit --tag "$IMAGE_NAME" --project "$PROJECT_ID" .

# 2. Deploy to Cloud Run
echo "Deploying to Cloud Run..."
gcloud run deploy "$SERVICE_NAME" \
  --image "$IMAGE_NAME" \
  --region "$REGION" \
  --project "$PROJECT_ID" \
  --service-account "$SERVICE_ACCOUNT" \
  --set-env-vars "DB_INSTANCE_CONNECTION_NAME=${DB_INSTANCE},YODAHUNTERS_DATABASE_NAME=${DB_NAME},YODAHUNTERS_DATABASE_USER=${DB_USER}" \
  --set-secrets "YODAHUNTERS_JWT_SECRET=yodahunters-jwt-secret:latest" \
  --allow-unauthenticated

echo "Deployment complete!"
gcloud run services describe "$SERVICE_NAME" --platform managed --region "$REGION" --project "$PROJECT_ID" --format='value(status.url)'

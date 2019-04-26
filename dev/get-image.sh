#!/usr/bin/env bash

set -e

if ! command -v jq >/dev/null 2>&1; then
    echo "Please install jq before continuing"
    exit 1
fi

echo Getting latest image

PIPELINE_ID=$(aws codepipeline list-pipeline-executions --pipeline-name kboom --max-items 1 | jq .pipelineExecutionSummaries[0].pipelineExecutionId -r)

REVISION_ID=$(aws codepipeline get-pipeline-execution --pipeline-name kboom --pipeline-execution-id $PIPELINE_ID | jq .pipelineExecution.artifactRevisions[0].revisionId -r | cut -c 1-7)

echo Got new revision: $REVISION_ID

sed -e "s/{{REVISION_ID}}/${REVISION_ID}/g" job.yaml.template > job.yaml

echo You can now do: kubectl apply -f job.yaml
import boto3
import json


def tagEcrRepo(event, context):
    ecrClient = boto3.client('ecr')

    messageJson = event["Records"][0]["Sns"]["Message"]

    message = json.loads(messageJson)
    repositoryName = message["ecr_repo_name"]
    tagToUpdate = message["ecr_tag_to_update"]
    tagToAdd = message["ecr_tag_to_add"]

    manifest = __getImageManifest(ecrClient, repositoryName, tagToUpdate)

    response = ecrClient.put_image(
        repositoryName=repositoryName,
        imageManifest=manifest,
        imageTag=tagToAdd,
    )

    return response


def __getImageManifest(ecrClient, repositoryName, tagToUpdate):
    response = ecrClient.batch_get_image(
        repositoryName=repositoryName,
        imageIds=[
            {
                'imageTag': tagToUpdate,
            },
        ],
        acceptedMediaTypes=[
            'application/vnd.docker.distribution.manifest.v2+json',
            'application/vnd.oci.image.manifest.v1+json',
        ]
    )

    __validateBatchGetImageResponse(response, tagToUpdate)

    return response["images"][0]["imageManifest"]


def __validateBatchGetImageResponse(response, tagToUpdate):
    if len(response["failures"]) > 0:
        exceptionMessage = "Failures trying to get image manifest with tag {tag}. {failures}".format(
            tag=tagToUpdate,
            failures=json.dumps(response["failures"]),
        )
        raise Exception(exceptionMessage)

    if len(response["images"]) != 1:
        exceptionMessage = "Got {number_of_images} images when looking for image with tag {tag}. Should be 1".format(
            number_of_images=len(response["images"]),
            tag=tagToUpdate,
        )
        raise Exception(exceptionMessage)

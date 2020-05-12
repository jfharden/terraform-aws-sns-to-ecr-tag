import json
import unittest
from unittest.mock import patch

from lambdas.tag_ecr_repo import tagEcrRepo
import botocore.session
from botocore.stub import Stubber


class TestTagECRRepo(unittest.TestCase):
    REPO_NAME = "testrepo"
    IMAGE_TAG = "latest"
    IMAGE_TAG_TO_ADD = "newtag"
    IMAGE_DIGEST = "sha256:12345abc"
    REGISTRY_ID = "12345678"
    IMAGE_MANIFEST = "fake_manifest"
    BATCH_GET_EXPECTED_PARAMS = {
        "repositoryName": REPO_NAME,
        "imageIds": [
            {
                "imageTag": IMAGE_TAG,
            },
        ],
        "acceptedMediaTypes": [
            "application/vnd.docker.distribution.manifest.v2+json",
            "application/vnd.oci.image.manifest.v1+json",
        ]
    }
    PUT_IMAGE_EXPECTED_PARAMS = {
        "repositoryName": REPO_NAME,
        "imageManifest": IMAGE_MANIFEST,
        "imageTag": IMAGE_TAG_TO_ADD,
    }
    PUT_IMAGE_SUCCESSFUL_RESPONSE = {
        "image": {
            "registryId": REGISTRY_ID,
            "repositoryName": REPO_NAME,
            "imageId": {
                "imageDigest": IMAGE_DIGEST,
                "imageTag": IMAGE_TAG_TO_ADD,
            },
            "imageManifest": IMAGE_MANIFEST,
        }
    }

    def test_successful_path(self):
        ecrClient = botocore.session.get_session().create_client("ecr")

        stubber = Stubber(ecrClient)

        self.__mock_batch_get_success(stubber)
        self.__mock_put_image_success(stubber)
        stubber.activate()

        with patch("boto3.client") as mockClient:
            mockClient.return_value = ecrClient

            putImageResponse = tagEcrRepo(self.__sns_event(), self.__lambda_context())

        self.assertEqual(putImageResponse, TestTagECRRepo.PUT_IMAGE_SUCCESSFUL_RESPONSE)

    def test_no_image_found(self):
        ecrClient = botocore.session.get_session().create_client("ecr")

        stubber = Stubber(ecrClient)

        self.__mock_batch_get_no_image(stubber)
        stubber.activate()

        with patch("boto3.client") as mockClient:
            mockClient.return_value = ecrClient

            with self.assertRaisesRegex(Exception, "^Failures trying to get image manifest with tag latest.*"):
                tagEcrRepo(self.__sns_event(), self.__lambda_context())

    def test_batch_get_failure_propogated(self):
        ecrClient = botocore.session.get_session().create_client("ecr")

        stubber = Stubber(ecrClient)

        self.__mock_batch_get_limit_exceeded(stubber)
        stubber.activate()

        with patch("boto3.client") as mockClient:
            mockClient.return_value = ecrClient

            # with self.assertRaisesRegex(Exception, "^Failures trying to get image manifest with tag latest.*"):
            with self.assertRaises(Exception) as cm:
                tagEcrRepo(self.__sns_event(), self.__lambda_context())

            self.assertEqual(cm.exception.response["Error"]["Code"], "RepositoryNotFoundException")

        pass

    def test_put_image_failure_propogated(self):
        ecrClient = botocore.session.get_session().create_client("ecr")

        stubber = Stubber(ecrClient)

        self.__mock_batch_get_success(stubber)
        self.__mock_put_image_limit_exceeded(stubber)
        stubber.activate()

        with patch("boto3.client") as mockClient:
            mockClient.return_value = ecrClient

            # with self.assertRaisesRegex(Exception, "^Failures trying to get image manifest with tag latest.*"):
            with self.assertRaises(Exception) as cm:
                tagEcrRepo(self.__sns_event(), self.__lambda_context())

            self.assertEqual(cm.exception.response["Error"]["Code"], "LimitExceededException")

    def __sns_event(self):
        with open("tests/fixtures/sns_event.json") as event_file:
            event = json.loads(event_file.read())

        return event

    def __lambda_context(self):
        # The context isn"t used at all
        return {}

    def __mock_batch_get_success(self, stubber):
        response = {
            "images": [
                {
                    "registryId": TestTagECRRepo.REGISTRY_ID,
                    "repositoryName": TestTagECRRepo.REPO_NAME,
                    "imageId": {
                        "imageDigest": TestTagECRRepo.IMAGE_DIGEST,
                        "imageTag": TestTagECRRepo.IMAGE_TAG,
                    },
                    "imageManifest": TestTagECRRepo.IMAGE_MANIFEST,
                },
            ],
            "failures": []
        }

        stubber.add_response("batch_get_image", response, TestTagECRRepo.BATCH_GET_EXPECTED_PARAMS)

    def __mock_batch_get_no_image(self, stubber):
        response = {
            "images": [],
            "failures": [
                {
                    "imageId": {
                        "imageTag": TestTagECRRepo.IMAGE_TAG,
                    },
                    "failureCode": "ImageNotFound",
                    "failureReason": "No image found"
                },
            ]
        }

        stubber.add_response("batch_get_image", response, TestTagECRRepo.BATCH_GET_EXPECTED_PARAMS)

    def __mock_batch_get_limit_exceeded(self, stubber):
        stubber.add_client_error(
            "batch_get_image",
            service_error_code="RepositoryNotFoundException",
            http_status_code=400,
            expected_params=TestTagECRRepo.BATCH_GET_EXPECTED_PARAMS
        )

    def __mock_put_image_success(self, stubber):
        stubber.add_response(
            "put_image",
            TestTagECRRepo.PUT_IMAGE_SUCCESSFUL_RESPONSE,
            TestTagECRRepo.PUT_IMAGE_EXPECTED_PARAMS
        )

    def __mock_put_image_limit_exceeded(self, stubber):
        stubber.add_client_error(
            "put_image",
            service_error_code="LimitExceededException",
            http_status_code=400,
            expected_params=TestTagECRRepo.PUT_IMAGE_EXPECTED_PARAMS
        )

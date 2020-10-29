# gcsdropbox

`gcsdropbox` is a tool for generating pre-signed URLs that allow you to upload and download files from a 
[Google Cloud Storage](https://cloud.google.com/storage) bucket. These URLs can be used on hosts and/or by third parties
without needing to give those hosts/users credentials for direct access.

It uses the [signBlob](https://cloud.google.com/iam/docs/reference/credentials/rest/v1/projects.serviceAccounts/signBlob)
api to sign the url using a service account's Google-managed key.

(The caller needs the `iam.serviceAccounts.signBlob` IAM permission on the service account.)

## Usage

### Uploading to GCS

To create a url for uploading:

    gcsdropbox -serviceAccount myserviceaccount@myproject.iam.gserviceaccount.com -bucket myBucket -name foo/bar.tar

To use the url to upload:

    curl -T localfileToUpload https://storage.googleapis.com/myBucket/foo/bar.tar?Expires=...

### Downloading from GCS

To create a URL for downloading:

    gcsdropbox -serviceAccount myserviceaccount@myproject.iam.gserviceaccount.com -bucket myBucket -name baz.zip -method GET

To use the url to download:

    curl -O https://storage.googleapis.com/myBucket/baz.zip?Expires=...

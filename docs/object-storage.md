# Object storage

GooseForum supports S3-compatible object storage through `minio-go`. Existing database-backed files remain readable after S3 is enabled; new post images are written to S3.

## Configuration

```toml
[storage]
driver = "s3"

[storage.s3]
endpoint = "https://objects.example.com"
bucket = "gooseforum"
accessKey = ""
secretKey = ""
sessionToken = ""
region = "us-east-1"
secure = true
pathStyle = true
```

The endpoint embedded in the presigned upload must be reachable by users' browsers. Keep the bucket private and grant the application credentials only the object operations required for the configured bucket.

## Browser upload CORS

Configure the bucket to accept `POST` requests from the forum origin. Replace the example origin with the exact public forum origin instead of allowing arbitrary origins.

```json
[
  {
    "AllowedOrigins": ["https://forum.example.com"],
    "AllowedMethods": ["POST"],
    "AllowedHeaders": ["*"],
    "ExposeHeaders": ["ETag"],
    "MaxAgeSeconds": 3600
  }
]
```

The browser requests a short-lived POST policy from GooseForum, uploads directly to the bucket, and then asks GooseForum to verify and publish the object. GooseForum checks ownership, object size, MIME type, and the decoded image header before making the file visible. Incomplete uploads are removed after two hours.

Pipeline for building/testing release
=====================================

Pipeline running at http://ci.starkandwayne.com:8080/pipelines/spruce-release

Setup pipeline in Concourse
---------------------------

```
fly -t sw c -c pipeline.yml --vars-from credentials.yml spruce-release --paused=false
```

Building/updating the base Docker image for tasks
-------------------------------------------------

Each task within all job build plans uses the same base Docker image for all dependencies. Using the same Docker image is a convenience. This section explains how to re-build and push it to Docker Hub.

All the resources used in the pipeline are shipped as independent Docker images and are outside the scope of this section.

```
fly -t sw configure \
  -c ci_image/pipeline.yml \
  --vars-from credentials.yml \
  spruce-release-image --paused=false
```

This will ask your targeted Concourse to pull down this project repository, and build the `ci_image/Dockerfile`, and push it to a Docker image on Docker Hub.

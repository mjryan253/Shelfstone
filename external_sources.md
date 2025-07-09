# External Sources and Considerations for "shelfstone" Renaming

This document lists items that might require attention outside of this repository due to the renaming from "libreplex" to "shelfstone".

## Docker Image

*   **Original Image Name**: `libreplex:latest`
*   **New Image Name**: `shelfstone:latest`

**Action Required**:
The Docker image needs to be rebuilt and pushed to a container registry under the new name `shelfstone:latest`. Any CI/CD pipelines or deployment scripts that referenced `libreplex:latest` will need to be updated to use `shelfstone:latest`.

## Docker Volume

*   **Original Volume Name**: `libreplex_cache`
*   **New Volume Name**: `shelfstone_cache`

**Action Required**:
When deploying the updated `docker-compose.yaml`, Docker will create a new volume named `shelfstone_cache`. If there is important data in the old `libreplex_cache` volume that needs to be preserved, it will need to be manually migrated to the new volume. The method for this depends on the Docker host environment.

**Note**: These considerations are based on the changes made in `docker-compose.yaml`. If "libreplex" was used in other contexts (e.g., cloud resource names, external monitoring dashboards, DNS entries), those will also need to be identified and updated. The current search within the repository did not find other instances.

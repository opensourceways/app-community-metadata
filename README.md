# README
This project aims to provide a simple and flexible feature which serves desired git repo files as restful
endpoints, the framework will keep track of remote git repo, watch the desired file or directory, and
notify plugins when and only when file(directory) changed.

# Feature
1. Pluggable watching repos and http endpoints.
2. Watching file and directory are both supported.
3. Https and ssh schema are both supported.

# Metadata list
This table below lists all of supported metadata and its original repo

| Content | Endpoint  | Source Repo | Folder(Files) |
|---|---|---|---|
| openEuler mirror lists  | https://api.osinfra.cn/meta/v1/metadata/openeuler/mirrors/all  |  https://gitee.com/openeuler/infrastructure |  ./mirrors |

# Quick Start

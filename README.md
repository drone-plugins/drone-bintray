# drone-bintray

[![Build Status](http://beta.drone.io/api/badges/drone-plugins/drone-bintray/status.svg)](http://beta.drone.io/drone-plugins/drone-bintray)
[![Coverage Status](https://aircover.co/badges/drone-plugins/drone-bintray/coverage.svg)](https://aircover.co/drone-plugins/drone-bintray)
[![](https://badge.imagelayers.io/plugins/drone-bintray:latest.svg)](https://imagelayers.io/?images=plugins/drone-bintray:latest 'Get your own badge on imagelayers.io')

Drone plugin to publish files and artifacts to Bintray

## Binary

Build the binary using `make`:

```
make deps build
```

### Example

```sh
./drone-bintray <<EOF
{
    "repo": {
        "clone_url": "git://github.com/drone/drone",
        "owner": "drone",
        "name": "drone",
        "full_name": "drone/drone"
    },
    "system": {
        "link_url": "https://beta.drone.io"
    },
    "build": {
        "number": 22,
        "status": "success",
        "started_at": 1421029603,
        "finished_at": 1421029813,
        "message": "Update the Readme",
        "author": "johnsmith",
        "author_email": "john.smith@gmail.com"
        "event": "push",
        "branch": "master",
        "commit": "436b7a6e2abaddfd35740527353e78a227ddcb2c",
        "ref": "refs/heads/master"
    },
    "workspace": {
        "root": "/drone/src",
        "path": "/drone/src/github.com/drone/drone"
    },
    "vargs": {
        "username": "octocat",
        "password": "pa$$word",
        "branch": "master"
        "artifacts": [{
            "file": "dist/myfile",
            "owner": "mycompany",
            "type": "executable",
            "repository": "reponame",
            "package": "pkgname"
            "version": "1.0",
            "target": "myfile",
            "publish": true,
            "override": true
        }, {
            "file": "dist/myfile.deb",
            "owner": "mycompany",
            "type": "Debian",
            "repository": "debian-repo",
            "package": "pkgname",
            "version": "1.0",
            "target": "myfile.deb",
            "distr": "ubuntu",
            "component": "main",
            "arch": [
                "amd64"
            ],
            "publish": true,
            "override": true
        }]
    }
}
EOF
```

## Docker

Build the container using `make`:

```
make deps docker
```

### Example

```sh
docker run -i plugins/drone-bintray <<EOF
{
    "repo": {
        "clone_url": "git://github.com/drone/drone",
        "owner": "drone",
        "name": "drone",
        "full_name": "drone/drone"
    },
    "system": {
        "link_url": "https://beta.drone.io"
    },
    "build": {
        "number": 22,
        "status": "success",
        "started_at": 1421029603,
        "finished_at": 1421029813,
        "message": "Update the Readme",
        "author": "johnsmith",
        "author_email": "john.smith@gmail.com"
        "event": "push",
        "branch": "master",
        "commit": "436b7a6e2abaddfd35740527353e78a227ddcb2c",
        "ref": "refs/heads/master"
    },
    "workspace": {
        "root": "/drone/src",
        "path": "/drone/src/github.com/drone/drone"
    },
    "vargs": {
        "username": "octocat",
        "password": "pa$$word",
        "branch": "master"
        "artifacts": [{
            "file": "dist/myfile",
            "owner": "mycompany",
            "type": "executable",
            "repository": "reponame",
            "package": "pkgname"
            "version": "1.0",
            "target": "myfile",
            "publish": true,
            "override": true
        }, {
            "file": "dist/myfile.deb",
            "owner": "mycompany",
            "type": "Debian",
            "repository": "debian-repo",
            "package": "pkgname",
            "version": "1.0",
            "target": "myfile.deb",
            "distr": "ubuntu",
            "component": "main",
            "arch": [
                "amd64"
            ],
            "publish": true,
            "override": true
        }]
    }
}
EOF
```

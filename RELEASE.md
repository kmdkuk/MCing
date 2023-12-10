Release procedure
=================

This document describes how to release a new version of MCing.

## Versioning

Follow [semantic versioning 2.0.0][semver] to choose a new version number.

## Bump version

1. Determine a new version number.  Export it as an environment variable:

    ```console
    $ VERSION=1.2.3
    ```

2. Make a new branch from the latest `main` with `git checkout -b release/v$VERSION`
3. Update version strings in `kustomization.yaml`.
5. Commit the change and create a pull request:

    ```console
    $ git commit -a -m "Bump version to $VERSION"
    $ git push origin release/v$VERSION
    ```

6. Merge the new pull request. Push tag automatically.

## (Option) Edit GitHub release page

You may edit [the GitHub release page](https://github.com/kmdkuk/MCing/releases/latest) to add further details.

[semver]: https://semver.org/spec/v2.0.0.html

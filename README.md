# versionupdater
Automatically update Gradle dependencies

## Installation ##
```shell
go get github.com/bashtian/versionupdater
```

## Usage ##
```shell
# run versionupdater in path of your app
versionupdater
```

Versions can be pinned. See https://github.com/hashicorp/go-version#version-constraints for valid conditions.

```
implementation 'com.google.android:flexbox:1.0.0' // 1.0.0

implementation 'com.google.android:flexbox:1.0.0' // >=1.0.0, <2.0.0
```

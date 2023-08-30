![YuDisk](./YuDisk.png)
### YuDisk is a simple golang module for work with yandex disk. This module tested with application folder only.

## `NewYuDisk(token string) (YDApi, error)`
***

Creates a new YDApi object with a token and a user agent `Mozilla/5.0 (Linux; Android 13; SM-A037U) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/112.0.0.0 Mobile Safari/537.36`.

### Example:
```go
package main

import (
	"github.com/pr0tr3x/YuDisk"
)

func main() {
	yd, _ := YuDisk.NewYuDisk("oauth_token")
}
```

## `Download(path string) ([]byte, error)`
***

Downloads a file from disk.

### Example:
```go
raw, err = yd.Download("app:/YuDisk.png")
if err != nil {
    fmt.Printf("%+v\n", err)
} else {
    os.WriteFile("./YuDisk_download.png", raw, 0644)
}
```

## `Upload(pathLocal string, pathCloud string, overwrite bool) (string, error)`
***

Loads the `pathLocal` file to the disk on the path `pathCloud`. Overwrites it if `overwrite` is true.  Returns id async operation.

### Example:
```go
operational_id, err := yd.Upload("./YuDisk.png", "app:/YuDisk.png", true)
if err != nil {
    fmt.Printf("%+v\n", err)
} else {
    fmt.Printf("%+v\n", operational_id)
}
```

## `GetResourceMeta(path string) (Resource, error)`
***

Gets resource metadata information (folder or file). Return `Resource{}` structure.

### Example:
```go
res, err := yd.GetResourceMeta("app:/")
if err != nil {
    fmt.Printf("%+v\n", err)
} else {
    fmt.Printf("%+v\n", res)
}
```

## `Delete(pathToCloudFile string, permanently bool) error`
***

Deletes a file on disk. If the `permanently` parameter is true, it does not move the file to the trash.

### Example:
```go
err = yd.Delete("app:/YuDisk.png", true)
if err != nil {
    fmt.Printf("%+v\n", err)
}
```

## `Move/Copy(from string, to string, overwrite bool) error`
***

Moving (Copie) the file `from` to `to`. If the `overwrite` parameter is set to true, the existing file is overwritten.

### Example:
```go
err := yd.Move("app:/Test", "app:/TestMove", false)
if err != nil {
    fmt.Printf("%+v\n", err)
}
//--------------------------
err := yd.Copy("app:/Test", "app:/TestMove", false)
if err != nil {
fmt.Printf("%+v\n", err)
}
```

## `MkDir(path string) error`
***

Creates new folder.

### Example:
```go
err := yd.MkDir("app:/Test")
if err != nil {
fmt.Printf("%+v\n", err)
}
```

## `OperationStatus(operationID string) (string, error)`
***

Gets the status of an asynchronous operation. Returns `success` string if operation completed.

### Example:
```go
status, err := yd.OperationStatus(status)
if err != nil {
    fmt.Printf("%+v\n", err)
} else {
    fmt.Printf("%+v\n", status)
}
```

## `SetUserAgent(agent string)`
***

Sets new user agent string.

### Example:
```go
yd.SetUserAgent("New_Agent")
```

## `SetProxy(proxyUrl string) error`
***

Sets proxy url for all requests.

### Example:
```go
yd.SetProxy("http://127.0.0.1:8080")
```
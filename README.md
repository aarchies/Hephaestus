



![Go version](https://img.shields.io/badge/go-%3E%3Dv1.18-9cf)



<div STYLE="page-break-after: always;"></div>

<p style="font-size: 20px"> 
    reusable util function library of go.
</p>


## <a href="https://www.golancet.cn/en/" target="_blank"> Website</a> | [简体中文](./README_zh-CN.md)

## Features

-   👏 Comprehensive, efficient and reusable.
-   🌍 Unit test for every exported function.

## Installation

### Note:

```go
go get github.com/aarchie/go-lib
```

## Usage


```go
import "github.com/aarchie/go-lib"
```

## Example

Here takes the string function Reverse (reverse order string) as an example, and the strutil package needs to be imported.

```go
package main

import (
    "fmt"
    "github.com/aarchie/go-lib"
)

func main() {
    s := "hello"
    rs := strutil.Reverse(s)
    fmt.Println(rs) //olleh
}
```

## Documentation

- [messagec](#user-content-algorithm)

- [Compare](#user-content-compare)

- [Concurrency](#user-content-concurrency)


<h3 id="algorithm">1.Message bus, can support kafka, rabbimq, chan . &nbsp; &nbsp; &nbsp; &nbsp;<a href="#index">index</a></h3>

```go
import "github.com/aarchie/go-lib/messagec"
```

#### Function list:



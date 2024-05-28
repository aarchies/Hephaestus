# Hephaestus

[![CI Status](https://github.com/ThreeDotsLabs/watermill/actions/workflows/master.yml/badge.svg)](https://github.com/ThreeDotsLabs/watermill/actions/workflows/master.yml)[![Go Reference](https://pkg.go.dev/badge/github.com/ThreeDotsLabs/watermill.svg)](https://pkg.go.dev/github.com/ThreeDotsLabs/watermill)[![Go Report Card](https://goreportcard.com/badge/github.com/ThreeDotsLabs/watermill)](https://goreportcard.com/report/github.com/ThreeDotsLabs/watermill)[![codecov](https://codecov.io/gh/ThreeDotsLabs/watermill/branch/master/graph/badge.svg)](https://codecov.io/gh/ThreeDotsLabs/watermill)[![codecov](https://img.shields.io/badge/go-%3E%3Dv1.18-9cf)](https://codecov.io/gh/ThreeDotsLabs/watermill)

<p style="font-size: 20px"> 
   Hephaestus is a Go library that provides a variety of util functions and message libraries for quickly building business applications. It is designed to quickly build event-driven applications based on an event-based message bus, which you can use with traditional pub/sub implementations like kafka or rabbitmq.
</p>



## <a href="https://www.golancet.cn/en/" target="_blank"> Website</a> | [简体中文](./README_zh-CN.md)

## Features

-   👏 **Easy** to understand.
-   🌍 **Universal** - Based on event-driven architecture, messaging, cqrs.
-   💅 **Flexible** with middlewares, plugins and Pub/Sub configurations.

## Getting Started

### Note:

```go
go get github.com/aarchies/Hephaestus
```

## Example



* Basic
    * [Your first app](examples/event_bus/) - **start here!**
    * 

## Documentation

- [messagec](#user-content-algorithm)

- [Compare](#user-content-compare)

- [Concurrency](#user-content-concurrency)


<h3 id="algorithm">1.Message bus, can support kafka, rabbimq, chan . &nbsp; &nbsp; &nbsp; &nbsp;<a href="#index">index</a></h3>

```go
import "github.com/aarchie/Hephaestus/messagec"
```

#### Function list:


# Hephaestus

<img align="right" width="200" src="https://encrypted-tbn0.gstatic.com/images?q=tbn:ANd9GcQkx4uHb-9HJqz0yi-buNBfTOvS1cbZZ2YVvQ&usqp=CAU">

[![Go Reference](https://pkg.go.dev/badge/github.com/ThreeDotsLabs/Hephaestus.svg)](https://pkg.go.dev/github.com/ThreeDotsLabs/Hephaestus)[![Go Report Card](https://goreportcard.com/badge/github.com/ThreeDotsLabs/Hephaestus)](https://goreportcard.com/report/github.com/ThreeDotsLabs/Hephaestus)[![codecov](https://codecov.io/gh/ThreeDotsLabs/Hephaestus/branch/master/graph/badge.svg)](https://codecov.io/gh/ThreeDotsLabs/Hephaestus)[![codecov](https://img.shields.io/badge/go-%3E%3Dv1.20-9cf)](https://codecov.io/gh/ThreeDotsLabs/Hephaestus)

Hephaestus is a Go library that provides a variety of util functions and message libraries for quickly building business applications. It is designed to quickly build event-driven applications based on an event-based message bus, which you can use with traditional pub/sub implementations like kafka or rabbitmq




## <a href="https://www.golancet.cn/en/" target="_blank"> Website</a> | [简体中文](./README_zh-CN.md)

## Features

-   👏 **Easy** to understand
-   🌍 **Universal** - Based on event-driven architecture, messaging, cqrs
-   💅 **Flexible** with middlewares, plugins and Pub/Sub configurations

## Getting Started

## Installation:

```go
go get github.com/aarchies/hephaestus
```

## Example

* Basic
    * [Eventbus](examples/event_bus/main.go) - **start here!**

## Documentation

- [Messagec](#user-content-algorithm)

- [Compare](#user-content-compare)

- [Concurrency](#user-content-concurrency)


<h3 id="algorithm">1.Message bus, can support kafka, rabbimq, chan . &nbsp; &nbsp; &nbsp; &nbsp;<a href="#index">index</a></h3>

```go
import "github.com/aarchie/hephaestus/messagec/cqrs"
```

#### Function list:


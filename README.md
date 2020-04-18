# go2gdc

## 简介

Go语言软件包[go2gdc][go2gdc]可以灵活筛选、下载、整合[NCI的GDC数据库][GDC]提供的多组学数据（包括TCGA）。

1. 无需任何编程或生信基础。
1. 大幅简化数据准备步骤。
1. 其设计框架及使用方法详见[说明文档][go2gdc book]。
1. 其编译后的[可执行文件][go2gdc binary file]适用于多种常见操作系统。

## 常见问题

- 若遇如下提示，说明当前网络连接速度过低，未能与[GDC服务器][GDC]保持连接：
    1. Post https://api.gdc.cancer.gov/...: net/http: TLS handshake timeout
    1. Get  https://api.gdc.cancer.gov/...: net/http: TLS handshake timeout
    1. Post https://api.gdc.cancer.gov/...: EOF


## Introduction

The [go2gdc][go2gdc] is an open-source (GPL-3) [Go][Go] package for flexible filtering, downloading and integrating multi-omics data from [NCI's Genomic Data Commons (GDC)][GDC].

1. No programming or bioinformatics skill is needed.
1. It simplifies the data preparation significantly.
1. Its [manual][go2gdc book] explains the details.
1. Its [binary distributions][go2gdc binary file] work on various operating system.

## FAQ

- Any of following messages means the speed of internet connection is too low to keep contact with [GDC server][GDC]:
    1. Post https://api.gdc.cancer.gov/...: net/http: TLS handshake timeout
    1. Get  https://api.gdc.cancer.gov/...: net/http: TLS handshake timeout
    1. Post https://api.gdc.cancer.gov/...: EOF

[GDC]: https://gdc.cancer.gov/
[Go]: https://golang.org/
[go2gdc]: https://github.com/weil911/go2gdc
[go2gdc source code]: https://github.com/weil911/go2gdc/tree/master/src
[go2gdc binary file]: https://github.com/weil911/go2gdc/tree/master/bin
[go2gdc book]: https://github.com/weil911/go2gdc/tree/master/doc/go2gdc.pdf


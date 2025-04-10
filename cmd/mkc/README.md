# Matroska Command

A CLI tool to inspect Matroska files. The tool is written in Go.

- [Introduction](#introduction)
- [Production readiness](#production-readiness)
- [Documents](#documents)

## Introduction

`mkc` stands for "Matroska Command". The name of the command follows the logic of the extension naming used for Matroska files. The most used extensions are `mkv` "Matroska Video", `mka` "Matroska Audio", and `mks` "Matroska Subtitle".

This library is based on the version of [RFC 9559][rfc9559] and the 10th iteration of [draft-ietf-cellar-codec][draft-ietf-cellar-codec-14]. None of these documents reached ["Internet Standard"](https://datatracker.ietf.org/html/rfc2026#section-4.1.3) status yet.

- RFC 9559 is still a [Proposed Standard](https://datatracker.ietf.org/doc/html/rfc2026#section-4.1.1)
- draft-ietf-cellar-codec is still an [Internet-Draft](https://datatracker.ietf.org/html/rfc2026#section-2.2).

The goal of this command line tool is to see how one would use the libraries provided by [github.com/coding-socks/matroska](https://github.com/coding-socks/matroska) and [github.com/coding-socks/ebml](https://github.com/coding-socks/ebml).

## Production readiness

**This project is still in alpha phase.** In this stage the public API can change between days.

Beta version will be considered when the feature set covers most of the documents the implementation is based on, and the public API is reached a mature state.

Stable version will be considered only if enough positive feedback is gathered to lock the public API and all document the implementation is based on became ["Internet Standard"](https://tools.ietf.org/html/rfc2026#section-4.1.3).

## Documents

### Official sites

- [libEBML](http://matroska-org.github.io/libebml/)
- [EBML Specification](https://matroska-org.github.io/libebml/specs.html)
- [Matroska](https://www.matroska.org/index.html)
- [Matroska Element Specification](https://matroska.org/technical/elements.html)
- [WebM](https://www.webmproject.org/)
- [WebM Container Guidelines](https://www.webmproject.org/docs/container/)

Huge thanks to the [Matroska.org](https://www.matroska.org/) for their work.

### IETF Documents

- [RFC 8794: Extensible Binary Meta Language][rfc8794]
- [RFC 9559: Matroska Media Container Format Specification][rfc9559]
- [draft-ietf-cellar-codec-14: Matroska Media Container Codec Specifications][draft-ietf-cellar-codec-14]

Huge thanks to the [IETF CELLAR Working Group](https://datatracker.ietf.org/wg/cellar/charter/) for their work.

[rfc8794]: https://tools.ietf.org/html/rfc8794
[rfc9559]: https://datatracker.ietf.org/doc/html/rfc9559
[draft-ietf-cellar-codec-14]: https://datatracker.ietf.org/doc/html/draft-ietf-cellar-codec-14

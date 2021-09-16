# Matroska Command

A CLI tool to inspect Matroska files. The tool is written in Go.

- [Introduction](#introduction)
- [Production readiness](#production-readiness)
- [Documents](#documents)

## Introduction

`mkc` stands for "Matroska Command". The name of the command follows the logic of the extension naming used for Matroska files. The most used extensions are `mkv` "Matroska Video", `mka` "Matroska Audio", and `mks` "Matroska Subtitle".

The library used by this command is based on the 7th iteration of [draft-ietf-cellar-matroska][draft-ietf-cellar-matroska-07] and the 6th iteration of [draft-ietf-cellar-codec][draft-ietf-cellar-codec-06]. None of these documents reached ["Internet Standard"](https://tools.ietf.org/html/rfc2026#section-4.1.3) status yet.

- draft-ietf-cellar-matroska is still an [Internet-Draft](https://tools.ietf.org/html/rfc2026#section-2.2).
- draft-ietf-cellar-codec is still an [Internet-Draft](https://tools.ietf.org/html/rfc2026#section-2.2).

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
- [draft-ietf-cellar-matroska-07: Matroska Media Container Format Specifications][draft-ietf-cellar-matroska-07]
- [draft-ietf-cellar-codec-06: Matroska Media Container Codec Specifications][draft-ietf-cellar-codec-06]

Huge thanks to the [IETF CELLAR Working Group](https://datatracker.ietf.org/wg/cellar/charter/) for their work.

[rfc8794]: https://tools.ietf.org/html/rfc8794
[draft-ietf-cellar-matroska-07]: https://www.ietf.org/archive/id/draft-ietf-cellar-matroska-07.html
[draft-ietf-cellar-codec-06]: https://www.ietf.org/archive/id/draft-ietf-cellar-codec-06.html

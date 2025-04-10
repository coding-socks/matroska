# Matroska

A Matroska parser written in Go.

- [Introduction](#introduction)
- [Production readiness](#production-readiness)
- [Documents](#documents)
- [Similar libraries](#similar-libraries)

## Introduction

> The Matroska Multimedia Container is a free and open container format, a file format that can hold an unlimited number of video, audio, picture, or subtitle tracks in one file.

Source: https://en.wikipedia.org/wiki/Matroska

This library is based on the version of [RFC 9559][rfc9559] and the 10th iteration of [draft-ietf-cellar-codec][draft-ietf-cellar-codec-14]. None of these documents reached ["Internet Standard"](https://datatracker.ietf.org/html/rfc2026#section-4.1.3) status yet.

- RFC 9559 is still a [Proposed Standard](https://datatracker.ietf.org/doc/html/rfc2026#section-4.1.1)
- draft-ietf-cellar-codec is still an [Internet-Draft](https://datatracker.ietf.org/html/rfc2026#section-2.2).

The goal of this project is to create an implementation based on these documents and during the implementation provide feedback.

## Production readiness

**This project is still in alpha phase.** In this stage the public API can change between days.

Beta version will be considered when the feature set covers most of the documents the implementation is based on, and the public API is reached a mature state.

Stable version will be considered only if enough positive feedback is gathered to lock the public API and all document the implementation is based on became ["Internet Standard"](https://datatracker.ietf.org/html/rfc2026#section-4.1.3).

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

- [RFC 9559: Matroska Media Container Format Specification][rfc9559]
- [draft-ietf-cellar-codec-14: Matroska Media Container Codec Specifications][draft-ietf-cellar-codec-14]

Huge thanks to the [IETF CELLAR Working Group](https://datatracker.ietf.org/wg/cellar/charter/) for their work.

## Inspiration

Inspiration for the implementation comes from the following places:

- https://pkg.go.dev/database/sql#Drivers
- https://pkg.go.dev/database/sql#Register
- https://pkg.go.dev/encoding/json#Decoder
- https://pkg.go.dev/golang.org/x/image/vp8#Decoder

## Similar libraries

Last updated: 2020-02-18

| URL                                                               | Status                      |
|-------------------------------------------------------------------|-----------------------------|
| https://github.com/at-wat/ebml-go                                 | In active development       |
| https://github.com/ebml-go/ebml + https://github.com/ebml-go/webm | Last updated on 25 Sep 2016 |
| https://github.com/ehmry/go-ebml                                  | Archived                    |
| https://github.com/jacereda/ebml                                  | Last updated on 10 Jan 2016 |
| https://github.com/mediocregopher/ebmlstream                      | Last updated on 15 Dec 2014 |
| https://github.com/pankrator/ebml-parser                          | Last updated on 24 Jun 2020 |
| https://github.com/pixelbender/go-matroska                        | Last updated on 29 Oct 2018 |
| https://github.com/pubblic/ebml                                   | Last updated on 12 Dec 2018 |
| https://github.com/quadrifoglio/go-mkv                            | Last updated on 20 Jun 2018 |
| https://github.com/rrerolle/ebml-go                               | Last updated on 1 Dec 2012  |
| https://github.com/remko/go-mkvparse                              | Last updated on 14 Jun 2020 |
| https://github.com/tpjg/ebml-go                                   | Last updated on 1 Dec 2012  |

[rfc9559]: https://datatracker.ietf.org/doc/html/rfc9559
[draft-ietf-cellar-codec-14]: https://datatracker.ietf.org/doc/html/draft-ietf-cellar-codec-14

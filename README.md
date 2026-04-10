# 🎵 GoTAU

[![CI (Go)](https://github.com/SladkyCitron/gotau/actions/workflows/ci.yml/badge.svg)](https://github.com/SladkyCitron/gotau/actions/workflows/ci.yml) [![GitHub license](https://img.shields.io/github/license/SladkyCitron/gotau)](LICENSE) [![Made in Slovakia](https://raw.githubusercontent.com/pedromxavier/flag-badges/refs/heads/main/badges/SK.svg)](https://www.youtube.com/watch?v=UqXJ0ktrmh0)

**GoTAU** (pronounced *Go UTAU*) is a work-in-progress modern UTAU-compatible singing voice synthesizer written in Go.
It's designed to be fast, modern, modular, and easy to use, while staying backwards compatible with the existing UTAU ecosystem.

This project is also my senior high school IT project.

Built with ❤️ for the vocal synth community.

## 🚧 Project Status

GoTAU is currently in early development.

Core systems are still being implemented, and many features are incomplete.
Expect bugs and breaking changes.

## ✨ Features

* Fast and efficient synthesis engine
* Cross-platform support
* Backwards compatibility with existing UST files and UTAU voicebanks
* Modular architecture for easy extension

### Planned Features

* Built-in resampler / concatenator
* Plugin support
* GUI

## ⚠️ Known Limitations

* No built-in resampler yet, users have to use an external one for now (I recommend [straycat-rs](https://github.com/UtaUtaUtau/straycat-rs))
* No built-in concatenator yet, users have to use an external one for now
* Only supports CV and VCV voicebanks

## 🐹 Why Go?

Most vocal synth frontends are written in C# or Java, and the underlying synthesis engines are often written in lower-level languages like C++.
GoTAU explores what a singing voice synthesizer would look like if it were entirely written in Go, leveraging Go's modern tooling, simplicity, concurrency model, and performance.

Also, I just really like Go and wanted to see if I could build a vocal synth with it!

## 📖 The name

The name "GoTAU" is a portmanteau of "Go" and "UTAU", pronounced "Go UTAU".

## ⚖️ License

MIT License (see [LICENSE](LICENSE))

## ❤️ Acknowledgements

* UTAU by Ameya/Ayame (飴屋／菖蒲)
* The Vocaloid & UTAU community

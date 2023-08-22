<p align="center"><a href="https://github.com/win32ss/supermium" target="_blank"><img width="100" src="https://github.com/adeii/supermium-portable/blob/master/res/papp.png"></a></p>

<p align="center">
Thanks to Crazy-Max for chrome portable source.
  <br /><a href="https://github.com/sponsors/crazy-max"><img src="https://img.shields.io/badge/sponsor-crazy--max-181717.svg?logo=github&style=flat-square" alt="Become a sponsor"></a>
  <a href="https://www.paypal.me/crazyws"><img src="https://img.shields.io/badge/donate-paypal-00457c.svg?logo=paypal&style=flat-square" alt="Donate Paypal"></a>
</p>

## Notice of Non-Affiliation and Disclaimer

Portapps is not affiliated, associated, authorized, endorsed by, or in any way officially connected with Ungoogled Chromium™, or any of its subsidiaries or its affiliates.

The official Supermium browser project can be found at https://github.com/win32ss/supermium.

The name Supermium as well as related names, marks, emblems and images are registered trademarks of their respective owners.

## About

Supermium portable app made with 🚀 [Portapps](https://github.com/portapps).<br />
No documentation was made. 

Supermium is Chromium-based browser for Windows Vista (with extended kernel), 7 and 8.x.
Author is win32ss. https://github.com/win32ss
Since Google cut support for Windows versions older than 10, project Supermium is reaction you are looking for. 

## License

MIT. See `LICENSE` for more details.<br />
Rocket icon credit to [Squid Ink](http://thesquid.ink).

## Building

You will need to install Go and Apache Ant and add binaries to system PATH.

mkdir C:\portapps-dev
cd C:\portapps-dev\
git clone https://github.com/adeii/supermium-portable
git clone https://github.com/portapps/portapps
cd supermium-portable\
ant release

You will find supermium-portable/bin/release/supermium-portable-win64-117*.7z upon completed building.

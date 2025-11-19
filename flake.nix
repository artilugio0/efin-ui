{
  description = "Go shell";

  inputs = {
    nixpkgs.url = "nixpkgs/nixos-25.05";
  };

  outputs = { nixpkgs, ... } :
    let
      system = "x86_64-linux";
      pkgs = nixpkgs.legacyPackages.${system};
    in
    {
      devShells.${system}.default = pkgs.mkShell {
        nativeBuildInputs = with pkgs; [
          go
          gopls
          gotools
          nodejs

          pkg-config
          gtk3
          webkitgtk_4_0  # or webkitgtk_4_0 / webkitgtk depending on Wails version
          gobject-introspection
          libsoup_3
          glib-networking  # for TLS
          pango
          harfbuzz
          cairo
          atkmm
          at-spi2-atk
          gdk-pixbuf
        ];

        shellHook = ''
          export PKG_CONFIG_PATH="${pkgs.glib.dev}/lib/pkgconfig\
:${pkgs.glib.dev}/lib/pkgconfig\
:${pkgs.gtk3.dev}/lib/pkgconfig\
:${pkgs.webkitgtk_4_0.dev}/lib/pkgconfig\
:${pkgs.harfbuzz.dev}/lib/pkgconfig\
:${pkgs.cairo.dev}/lib/pkgconfig\
:${pkgs.libsoup_3.dev}/lib/pkgconfig\
:${pkgs.pango.dev}/lib/pkgconfig\
:${pkgs.gdk-pixbuf.dev}/lib/pkgconfig\
:${pkgs.atkmm.dev}/lib/pkgconfig\
:${pkgs.at-spi2-atk.dev}/lib/pkgconfig\
:${pkgs.gobject-introspection.dev}/lib/pkgconfig:$PKG_CONFIG_PATH"

          export GOBIN=$HOME/go/bin
          export PATH=$PATH:$GOBIN
        '';
      };
    };
}

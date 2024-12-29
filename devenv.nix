{ pkgs, lib, config, inputs, ... }:

{
  packages = [ pkgs.jq pkgs.git ];

  env = {
    SCEP_CHALLENGE = "monimentormdm";
  };

  scripts = {
   init.exec = ''
    ./scepserver-darwin-arm64 ca -init -country=NL -common_name="Monimentor SCEP CA" -organization="Qabam B.V." -organizational_unit="Monimentor"
   '';

   serve.exec = ''
    ./scepserver-darwin-arm64 -allowrenew 0 -challenge $SCEP_CHALLENGE -debug -port 7048
   '';
  };

  languages = {
    go = {
      enable = true;
    };
  };
}

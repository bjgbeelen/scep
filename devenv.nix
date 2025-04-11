{ pkgs, lib, config, inputs, ... }:

{
  packages = (with pkgs; [ jq git
  (google-cloud-sdk.withExtraComponents([google-cloud-sdk.components.gke-gcloud-auth-plugin]))
  ]);

  env = {
    # SCEP_CHALLENGE = "monimentormdm";
    SCEP_CHALLENGE_URL = "http://my.localhost.com/v1/device/challenge";
  };

  scripts = {
    init.exec = ''
      ./scepserver-darwin-arm64 ca -init -country=NL -common_name="Monimentor SCEP CA" -organization="Qabam B.V." -organizational_unit="Monimentor"
    '';

    serve.exec = ''
      ./scepserver-darwin-arm64 -allowrenew 0 -debug -port 7048
    '';

    get-project-id.exec = ''
      gcloud projects list --filter "name:$1" --format="value(project_id)"
    '';

    short-sha.exec = ''
      git rev-parse --short HEAD
    '';

    repo-name.exec = ''
      echo $(basename $(pwd))
    '';

    image-name.exec = ''
      ops_project_id=$(get-project-id monimentor-ops)
      echo europe-west4-docker.pkg.dev/$ops_project_id/docker/monimentor-$(repo-name)
    '';

    full-image-name.exec = ''
      echo "$(image-name):$(short-sha)"
    '';

    build.exec = ''
      set -e
      env=$1
      GOOS=linux GOARCH=amd64 go build -ldflags "-X main.version=$(short-sha)" -o scepserver-linux-amd64 ./cmd/scepserver
      gcloud auth configure-docker europe-west4-docker.pkg.dev
      full_image_name=$(full-image-name)
      docker buildx build --platform linux/amd64 --build-arg GIT_SHA=$(short-sha) \
       -t $full_image_name --load .
      for additional_tag in "$@"
      do
        docker tag $full_image_name "$(image-name):$additional_tag"
      done
      docker push -a $(image-name)
    '';
  };

  languages = {
    go = {
      enable = true;
    };
  };
}

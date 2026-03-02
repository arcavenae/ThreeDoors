class Threedoors < Formula
  desc "Three Doors - radical task management showing only 3 tasks at a time"
  homepage "https://github.com/arcaven/ThreeDoors"
  version "VERSION_PLACEHOLDER"
  license "MIT"

  on_arm do
    url "https://github.com/arcaven/ThreeDoors/releases/download/TAG_PLACEHOLDER/threedoors-darwin-arm64"
    sha256 "SHA256_ARM64_PLACEHOLDER"
  end

  on_intel do
    url "https://github.com/arcaven/ThreeDoors/releases/download/TAG_PLACEHOLDER/threedoors-darwin-amd64"
    sha256 "SHA256_AMD64_PLACEHOLDER"
  end

  def install
    binary_name = Hardware::CPU.arm? ? "threedoors-darwin-arm64" : "threedoors-darwin-amd64"
    bin.install binary_name => "threedoors"
  end

  test do
    assert_match "ThreeDoors", shell_output("#{bin}/threedoors --version 2>&1", 0)
  end
end

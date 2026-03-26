# Draft Homebrew Formula for ThreeDoors
# This is a SOURCE BUILD formula suitable for homebrew-core submission.
# For the custom tap, GoReleaser will auto-generate a binary-distribution formula.
#
# Usage (homebrew-core submission):
#   1. Replace VERSION and SHA256 with actual values
#   2. Place at Formula/t/threedoors.rb in homebrew-core fork
#   3. Run: brew audit --strict --new --online threedoors
#   4. Run: brew install --build-from-source threedoors
#   5. Run: brew test threedoors
#   6. Submit PR to Homebrew/homebrew-core

class Threedoors < Formula
  desc "TUI task manager that reduces decision friction by showing only three tasks"
  homepage "https://github.com/arcavenae/ThreeDoors"
  url "https://github.com/arcavenae/ThreeDoors/archive/refs/tags/v1.0.0.tar.gz"
  sha256 "REPLACE_WITH_ACTUAL_SHA256"
  license "MIT"
  head "https://github.com/arcavenae/ThreeDoors.git", branch: "main"

  depends_on "go" => :build

  def install
    ldflags = %W[
      -s -w
      -X main.version=#{version}
      -X github.com/arcavenae/ThreeDoors/internal/cli.Version=#{version}
      -X github.com/arcavenae/ThreeDoors/internal/cli.Commit=HEAD
      -X github.com/arcavenae/ThreeDoors/internal/cli.BuildDate=#{time.iso8601}
    ]
    system "go", "build", *std_go_args(ldflags:), "./cmd/threedoors"
  end

  test do
    # Verify version output
    assert_match version.to_s, shell_output("#{bin}/threedoors --version")

    # Verify help output
    assert_match "ThreeDoors", shell_output("#{bin}/threedoors --help")

    # Verify the binary can initialize without a tasks file
    # (should exit gracefully, not crash)
    output = shell_output("#{bin}/threedoors --version 2>&1")
    assert_match(/\d+\.\d+\.\d+/, output)
  end
end

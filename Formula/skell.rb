# To use this formula, add the tap first:
#   brew tap aminmesbahi/tap
#   brew install skell
#
# Or install directly:
#   brew install aminmesbahi/tap/skell
#
# This file belongs in a GitHub repository named:
#   aminmesbahi/homebrew-tap  (at Formula/skell.rb)

class Skell < Formula
  desc "Cross-platform skill package manager for Agent Skills (SKILL.md)"
  homepage "https://github.com/aminmesbahi/skell"
  version "0.1.6" # update on each release

  on_macos do
    on_arm do
      url "https://github.com/aminmesbahi/skell/releases/download/v#{version}/skell_#{version}_darwin_arm64.tar.gz"
      # sha256 "<darwin_arm64_sha256>"  # fill in after generating a release
    end
    on_intel do
      url "https://github.com/aminmesbahi/skell/releases/download/v#{version}/skell_#{version}_darwin_amd64.tar.gz"
      # sha256 "<darwin_amd64_sha256>"  # fill in after generating a release
    end
  end

  on_linux do
    on_arm do
      url "https://github.com/aminmesbahi/skell/releases/download/v#{version}/skell_#{version}_linux_arm64.tar.gz"
      # sha256 "<linux_arm64_sha256>"   # fill in after generating a release
    end
    on_intel do
      url "https://github.com/aminmesbahi/skell/releases/download/v#{version}/skell_#{version}_linux_amd64.tar.gz"
      # sha256 "<linux_amd64_sha256>"   # fill in after generating a release
    end
  end

  license "MIT"

  def install
    bin.install "skell"
  end

  test do
    assert_match version.to_s, shell_output("#{bin}/skell --version")
  end
end

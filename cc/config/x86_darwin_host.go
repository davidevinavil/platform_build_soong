// Copyright 2016 Google Inc. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package config

import (
	"fmt"
	"os/exec"
	"strings"

	"android/soong/android"
)

var (
	darwinCflags = []string{
		"-fno-exceptions", // from build/core/combo/select.mk
		"-Wno-multichar",  // from build/core/combo/select.mk

		"-fdiagnostics-color",

		"-fPIC",
		"-funwind-tables",

		// Workaround differences in inttypes.h between host and target.
		//See bug 12708004.
		"-D__STDC_FORMAT_MACROS",
		"-D__STDC_CONSTANT_MACROS",

		// HOST_RELEASE_CFLAGS
		"-O2", // from build/core/combo/select.mk
		"-g",  // from build/core/combo/select.mk
		"-fno-strict-aliasing", // from build/core/combo/select.mk
		"-isysroot ${macSdkRoot}",
		"-mmacosx-version-min=${macSdkVersion}",
		"-DMACOSX_DEPLOYMENT_TARGET=${macSdkVersion}",
	}

	darwinLdflags = []string{
		"-isysroot ${macSdkRoot}",
		"-Wl,-syslibroot,${macSdkRoot}",
		"-mmacosx-version-min=${macSdkVersion}",
	}

	// Extended cflags
	darwinX86Cflags = []string{
		"-m32",
	}

	darwinX8664Cflags = []string{
		"-m64",
	}

	darwinX86Ldflags = []string{
		"-m32",
	}

	darwinX8664Ldflags = []string{
		"-m64",
	}

	darwinClangCflags = append(ClangFilterUnknownCflags(darwinCflags), []string{
		"-integrated-as",
		"-fstack-protector-strong",
	}...)

	darwinX86ClangCflags = append(ClangFilterUnknownCflags(darwinX86Cflags), []string{
		"-msse3",
	}...)

	darwinClangLdflags = ClangFilterUnknownCflags(darwinLdflags)

	darwinX86ClangLdflags = ClangFilterUnknownCflags(darwinX86Ldflags)

	darwinX8664ClangLdflags = ClangFilterUnknownCflags(darwinX8664Ldflags)

	darwinSupportedSdkVersions = []string{
		"10.8",
		"10.9",
		"10.10",
		"10.11",
	}

	darwinAvailableLibraries = append(
		addPrefix([]string{
			"c",
			"dl",
			"m",
			"ncurses",
			"objc",
			"pthread",
			"z",
		}, "-l"),
		"-framework CoreFoundation",
		"-framework IOKit",
	)
)

const (
	darwinGccVersion = "4.2.1"
)

func init() {
	pctx.VariableFunc("macSdkPath", func(config interface{}) (string, error) {
		bytes, err := exec.Command("xcode-select", "--print-path").Output()
		return strings.TrimSpace(string(bytes)), err
	})
	pctx.VariableFunc("macSdkRoot", func(config interface{}) (string, error) {
		return xcrunSdk(config.(android.Config), "--show-sdk-path")
	})
	pctx.StaticVariable("macSdkVersion", darwinSupportedSdkVersions[0])
	pctx.VariableFunc("MacArPath", func(config interface{}) (string, error) {
		bytes, err := exec.Command("xcrun", "--find", "ar").Output()
		return strings.TrimSpace(string(bytes)), err
	})

	pctx.VariableFunc("MacStripPath", func(config interface{}) (string, error) {
		bytes, err := exec.Command("xcrun", "--find", "strip").Output()
		return strings.TrimSpace(string(bytes)), err
	})

	pctx.StaticVariable("DarwinGccVersion", darwinGccVersion)
	pctx.SourcePathVariable("DarwinGccRoot",
		"prebuilts/gcc/${HostPrebuiltTag}/host/i686-apple-darwin-${DarwinGccVersion}")

	pctx.StaticVariable("DarwinGccTriple", "i686-apple-darwin11")

	pctx.StaticVariable("DarwinCflags", strings.Join(darwinCflags, " "))
	pctx.StaticVariable("DarwinLdflags", strings.Join(darwinLdflags, " "))

	pctx.StaticVariable("DarwinClangCflags", strings.Join(darwinClangCflags, " "))
	pctx.StaticVariable("DarwinClangLdflags", strings.Join(darwinClangLdflags, " "))

	// Extended cflags
	pctx.StaticVariable("DarwinX86Cflags", strings.Join(darwinX86Cflags, " "))
	pctx.StaticVariable("DarwinX8664Cflags", strings.Join(darwinX8664Cflags, " "))
	pctx.StaticVariable("DarwinX86Ldflags", strings.Join(darwinX86Ldflags, " "))
	pctx.StaticVariable("DarwinX8664Ldflags", strings.Join(darwinX8664Ldflags, " "))

	pctx.StaticVariable("DarwinX86ClangCflags", strings.Join(darwinX86ClangCflags, " "))
	pctx.StaticVariable("DarwinX8664ClangCflags",
		strings.Join(ClangFilterUnknownCflags(darwinX8664Cflags), " "))
	pctx.StaticVariable("DarwinX86ClangLdflags", strings.Join(darwinX86ClangLdflags, " "))
	pctx.StaticVariable("DarwinX8664ClangLdflags", strings.Join(darwinX8664ClangLdflags, " "))
}

func xcrunSdk(config android.Config, arg string) (string, error) {
	if selected := config.Getenv("MAC_SDK_VERSION"); selected != "" {
		if !inList(selected, darwinSupportedSdkVersions) {
			return "", fmt.Errorf("MAC_SDK_VERSION %s isn't supported: %q", selected, darwinSupportedSdkVersions)
		}

		bytes, err := exec.Command("xcrun", "--sdk", "macosx"+selected, arg).Output()
		if err == nil {
			return strings.TrimSpace(string(bytes)), err
		}
		return "", fmt.Errorf("MAC_SDK_VERSION %s is not installed", selected)
	}

	for _, sdk := range darwinSupportedSdkVersions {
		bytes, err := exec.Command("xcrun", "--sdk", "macosx"+sdk, arg).Output()
		if err == nil {
			return strings.TrimSpace(string(bytes)), err
		}
	}
	return "", fmt.Errorf("Could not find a supported mac sdk: %q", darwinSupportedSdkVersions)
}

type toolchainDarwin struct {
	cFlags, ldFlags string
}

type toolchainDarwinX86 struct {
	toolchain32Bit
	toolchainDarwin
}

type toolchainDarwinX8664 struct {
	toolchain64Bit
	toolchainDarwin
}

func (t *toolchainDarwinX86) Name() string {
	return "x86"
}

func (t *toolchainDarwinX8664) Name() string {
	return "x86_64"
}

func (t *toolchainDarwin) GccRoot() string {
	return "${config.DarwinGccRoot}"
}

func (t *toolchainDarwin) GccTriple() string {
	return "${config.DarwinGccTriple}"
}

func (t *toolchainDarwin) GccVersion() string {
	return darwinGccVersion
}

func (t *toolchainDarwin) Cflags() string {
	return "${config.DarwinCflags} ${config.DarwinX86Cflags}"
}

func (t *toolchainDarwinX8664) Cflags() string {
	return "${config.DarwinCflags} ${config.DarwinX8664Cflags}"
}

func (t *toolchainDarwin) Cppflags() string {
	return ""
}

func (t *toolchainDarwinX86) Ldflags() string {
	return "${config.DarwinLdflags} ${config.DarwinX86Ldflags}"
}

func (t *toolchainDarwinX8664) Ldflags() string {
	return "${config.DarwinLdflags} ${config.DarwinX8664Ldflags}"
}

func (t *toolchainDarwin) IncludeFlags() string {
	return ""
}

func (t *toolchainDarwinX86) ClangTriple() string {
	return "i686-apple-darwin"
}

func (t *toolchainDarwinX86) ClangCflags() string {
	return "${config.DarwinClangCflags} ${config.DarwinX86ClangCflags}"
}

func (t *toolchainDarwinX8664) ClangTriple() string {
	return "x86_64-apple-darwin"
}

func (t *toolchainDarwinX8664) ClangCflags() string {
	return "${config.DarwinClangCflags} ${config.DarwinX8664ClangCflags}"
}

func (t *toolchainDarwin) ClangCppflags() string {
	return ""
}

func (t *toolchainDarwinX86) ClangLdflags() string {
	return "${config.DarwinClangLdflags} ${config.DarwinX86ClangLdflags}"
}

func (t *toolchainDarwinX8664) ClangLdflags() string {
	return "${config.DarwinClangLdflags} ${config.DarwinX8664ClangLdflags}"
}

func (t *toolchainDarwin) ShlibSuffix() string {
	return ".dylib"
}

func (t *toolchainDarwin) AvailableLibraries() []string {
	return darwinAvailableLibraries
}

var toolchainDarwinX86Singleton Toolchain = &toolchainDarwinX86{}
var toolchainDarwinX8664Singleton Toolchain = &toolchainDarwinX8664{}

func darwinX86ToolchainFactory(arch android.Arch) Toolchain {
	return toolchainDarwinX86Singleton
}

func darwinX8664ToolchainFactory(arch android.Arch) Toolchain {
	return toolchainDarwinX8664Singleton
}

func init() {
	registerToolchainFactory(android.Darwin, android.X86, darwinX86ToolchainFactory)
	registerToolchainFactory(android.Darwin, android.X86_64, darwinX8664ToolchainFactory)
}

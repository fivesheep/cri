/*
Copyright 2017 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package io

import (
	"bytes"
	"io/ioutil"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	runtime "k8s.io/kubernetes/pkg/kubelet/apis/cri/runtime/v1alpha2"

	cioutil "github.com/containerd/cri/pkg/ioutil"
)

func TestRedirectLogs(t *testing.T) {
	for desc, test := range map[string]struct {
		input   string
		stream  StreamType
		tag     []runtime.LogTag
		content []string
	}{
		"stdout log": {
			input:  "test stdout log 1\ntest stdout log 2",
			stream: Stdout,
			tag: []runtime.LogTag{
				runtime.LogTagFull,
				runtime.LogTagFull,
			},
			content: []string{
				"test stdout log 1",
				"test stdout log 2",
			},
		},
		"stderr log": {
			input:  "test stderr log 1\ntest stderr log 2",
			stream: Stderr,
			tag: []runtime.LogTag{
				runtime.LogTagFull,
				runtime.LogTagFull,
			},
			content: []string{
				"test stderr log 1",
				"test stderr log 2",
			},
		},
		"long log": {
			input:  strings.Repeat("a", 2*bufSize+10) + "\n",
			stream: Stdout,
			tag: []runtime.LogTag{
				runtime.LogTagPartial,
				runtime.LogTagPartial,
				runtime.LogTagFull,
			},
			content: []string{
				strings.Repeat("a", bufSize),
				strings.Repeat("a", bufSize),
				strings.Repeat("a", 10),
			},
		},
	} {
		t.Logf("TestCase %q", desc)
		rc := ioutil.NopCloser(strings.NewReader(test.input))
		buf := bytes.NewBuffer(nil)
		wc := cioutil.NewNopWriteCloser(buf)
		redirectLogs("test-path", rc, wc, test.stream)
		output := buf.String()
		lines := strings.Split(output, "\n")
		lines = lines[:len(lines)-1] // Discard empty string after last \n
		assert.Len(t, lines, len(test.content))
		for i := range lines {
			fields := strings.SplitN(lines[i], string([]byte{delimiter}), 4)
			require.Len(t, fields, 4)
			_, err := time.Parse(timestampFormat, fields[0])
			assert.NoError(t, err)
			assert.EqualValues(t, test.stream, fields[1])
			assert.Equal(t, string(test.tag[i]), fields[2])
			assert.Equal(t, test.content[i], fields[3])
		}
	}
}

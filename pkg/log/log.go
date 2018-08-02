/*
 * Copyright 2018 The Service Manager Authors
 *
 *    Licensed under the Apache License, Version 2.0 (the "License");
 *    you may not use this file except in compliance with the License.
 *    You may obtain a copy of the License at
 *
 *        http://www.apache.org/licenses/LICENSE-2.0
 *
 *    Unless required by applicable law or agreed to in writing, software
 *    distributed under the License is distributed on an "AS IS" BASIS,
 *    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *    See the License for the specific language governing permissions and
 *    limitations under the License.
 */

// Package log contains logic for setting up logging for SM
package log

import (
	"fmt"

	"github.com/sirupsen/logrus"
)

// Settings type to be loaded from the environment
type Settings struct {
	Level  string
	Format string
}

// Validate validates the logging settings
func (s *Settings) Validate() error {
	if len(s.Level) == 0 {
		return fmt.Errorf("validate Settings: LogLevel missing")
	}
	if len(s.Format) == 0 {
		return fmt.Errorf("validate Settings: LogFormat missing")
	}
	return nil
}

var supportedFormatters = map[string]logrus.Formatter{
	"json": &logrus.JSONFormatter{},
	"text": &logrus.TextFormatter{},
}

// SetupLogging configures logrus logging using the provided settings
func SetupLogging(settings Settings) {
	level, err := logrus.ParseLevel(settings.Level)
	if err != nil {
		panic(fmt.Sprintf("Could not parse log level configuration: %s", err.Error()))
	}
	logrus.SetLevel(level)
	formatter, ok := supportedFormatters[settings.Format]
	if !ok {
		panic(fmt.Sprintf("Invalid log format: %s", settings.Format))
	}
	logrus.SetFormatter(formatter)
}

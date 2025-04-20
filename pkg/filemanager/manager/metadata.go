package manager

import (
	"context"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"github.com/cloudreve/Cloudreve/v4/application/constants"
	"github.com/cloudreve/Cloudreve/v4/application/dependency"
	"github.com/cloudreve/Cloudreve/v4/pkg/filemanager/fs"
	"github.com/cloudreve/Cloudreve/v4/pkg/filemanager/fs/dbfs"
	"github.com/cloudreve/Cloudreve/v4/pkg/hashid"
	"github.com/cloudreve/Cloudreve/v4/pkg/serializer"
	"github.com/go-playground/validator/v10"
	"strings"
)

type (
	metadataValidator func(ctx context.Context, m *manager, patch *fs.MetadataPatch) error
)

const (
	wildcardMetadataKey      = "*"
	customizeMetadataSuffix  = "customize"
	tagMetadataSuffix        = "tag"
	iconColorMetadataKey     = customizeMetadataSuffix + ":icon_color"
	emojiIconMetadataKey     = customizeMetadataSuffix + ":emoji"
	shareOwnerMetadataKey    = dbfs.MetadataSysPrefix + "shared_owner"
	shareRedirectMetadataKey = dbfs.MetadataSysPrefix + "shared_redirect"
)

var (
	validate = validator.New()

	lastEmojiHash = ""
	emojiPresets  = map[string]struct{}{}

	// validateColor validates a color value
	validateColor = func(optional bool) metadataValidator {
		return func(ctx context.Context, m *manager, patch *fs.MetadataPatch) error {
			if patch.Remove {
				return nil
			}

			tag := "omitempty,iscolor"
			if !optional {
				tag = "required,iscolor"
			}

			res := validate.Var(patch.Value, tag)
			if res != nil {
				return fmt.Errorf("invalid color: %w", res)
			}

			return nil
		}
	}
	validators = map[string]map[string]metadataValidator{
		"sys": {
			wildcardMetadataKey: func(ctx context.Context, m *manager, patch *fs.MetadataPatch) error {
				if patch.Remove {
					return fmt.Errorf("cannot remove system metadata")
				}

				dep := dependency.FromContext(ctx)
				// Validate share owner is valid hashid
				if patch.Key == shareOwnerMetadataKey {
					hasher := dep.HashIDEncoder()
					_, err := hasher.Decode(patch.Value, hashid.UserID)
					if err != nil {
						return fmt.Errorf("invalid share owner: %w", err)
					}

					return nil
				}

				// Validate share redirect uri is valid share uri
				if patch.Key == shareRedirectMetadataKey {
					uri, err := fs.NewUriFromString(patch.Value)
					if err != nil || uri.FileSystem() != constants.FileSystemShare {
						return fmt.Errorf("invalid redirect uri: %w", err)
					}

					return nil
				}

				return fmt.Errorf("unsupported system metadata key: %s", patch.Key)
			},
		},
		"dav": {},
		customizeMetadataSuffix: {
			iconColorMetadataKey: validateColor(false),
			emojiIconMetadataKey: func(ctx context.Context, m *manager, patch *fs.MetadataPatch) error {
				if patch.Remove {
					return nil
				}

				// Validate if patched emoji is within preset list.
				emojis := m.settings.EmojiPresets(ctx)
				current := fmt.Sprintf("%x", (sha1.Sum([]byte(emojis))))
				if current != lastEmojiHash {
					presets := make(map[string][]string)
					if err := json.Unmarshal([]byte(emojis), &presets); err != nil {
						return fmt.Errorf("failed to read emoji setting: %w", err)
					}

					emojiPresets = make(map[string]struct{})
					for _, v := range presets {
						for _, emoji := range v {
							emojiPresets[emoji] = struct{}{}
						}
					}
				}

				if _, ok := emojiPresets[patch.Value]; !ok {
					return fmt.Errorf("unsupported emoji")
				}
				return nil
			},
		},
		tagMetadataSuffix: {
			wildcardMetadataKey: func(ctx context.Context, m *manager, patch *fs.MetadataPatch) error {
				if err := validateColor(true)(ctx, m, patch); err != nil {
					return err
				}

				if patch.Key == tagMetadataSuffix+":" {
					return fmt.Errorf("invalid metadata key")
				}

				return nil
			},
		},
	}
)

func (m *manager) PatchMedata(ctx context.Context, path []*fs.URI, data ...fs.MetadataPatch) error {
	if err := m.validateMetadata(ctx, data...); err != nil {
		return err
	}

	return m.fs.PatchMetadata(ctx, path, data...)
}

func (m *manager) validateMetadata(ctx context.Context, data ...fs.MetadataPatch) error {
	for _, patch := range data {
		category := strings.Split(patch.Key, ":")
		if len(category) < 2 {
			return serializer.NewError(serializer.CodeParamErr, "Invalid metadata key", nil)
		}

		categoryValidators, ok := validators[category[0]]
		if !ok {
			return serializer.NewError(serializer.CodeParamErr, "Invalid metadata key",
				fmt.Errorf("unknown category: %s", category[0]))
		}

		// Explicit validators
		if v, ok := categoryValidators[patch.Key]; ok {
			if err := v(ctx, m, &patch); err != nil {
				return serializer.NewError(serializer.CodeParamErr, "Invalid metadata patch", err)
			}
		}

		// Wildcard validators
		if v, ok := categoryValidators[wildcardMetadataKey]; ok {
			if err := v(ctx, m, &patch); err != nil {
				return serializer.NewError(serializer.CodeParamErr, "Invalid metadata patch", err)
			}
		}
	}

	return nil
}

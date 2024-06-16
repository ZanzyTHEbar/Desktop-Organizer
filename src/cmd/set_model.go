package cmd

import (
	"desktop-cleaner/api"
	"desktop-cleaner/auth"
	"desktop-cleaner/internal"
	"desktop-cleaner/lib"
	"desktop-cleaner/term"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

var provider string

func init() {
	RootCmd.AddCommand(modelsSetCmd)

	modelsSetCmd.AddCommand(defaultModelSetCmd)

}

var modelsSetCmd = &cobra.Command{
	Use:   "set-model [model-set-or-role-or-setting] [property-or-value] [value]",
	Short: "Update current plan model settings",
	Run:   modelsSet,
	Args:  cobra.MaximumNArgs(3),
}

var defaultModelSetCmd = &cobra.Command{
	Use:   "default [model-set-or-role-or-setting] [property-or-value] [value]",
	Short: "Update org-wide default model settings",
	Run:   defaultModelsSet,
	Args:  cobra.MaximumNArgs(3),
}

func modelsSet(cmd *cobra.Command, args []string) {
	auth.MustResolveAuthWithOrg()
	lib.MustResolveProject()

	term.StartSpinner("")
	originalSettings, apiErr := api.Client.GetSettings(lib.CurrentPlanId, lib.CurrentBranch)
	term.StopSpinner()

	if apiErr != nil {
		term.OutputErrorAndExit("Error getting current settings: %v", apiErr)
		return
	}

	settings := updateModelSettings(args, originalSettings)

	if settings == nil {
		return
	}

	term.StartSpinner("")
	res, apiErr := api.Client.UpdateSettings(
		lib.CurrentPlanId,
		lib.CurrentBranch,
		internal.UpdateSettingsRequest{
			Settings: settings,
		})
	term.StopSpinner()

	if apiErr != nil {
		term.OutputErrorAndExit("Error updating settings: %v", apiErr)
		return
	}

	fmt.Println(res.Msg)
	fmt.Println()
	term.PrintCmds("", "models", "set-model default", "log")
}

func defaultModelsSet(cmd *cobra.Command, args []string) {
	auth.MustResolveAuthWithOrg()

	term.StartSpinner("")
	originalSettings, apiErr := api.Client.GetOrgDefaultSettings()
	term.StopSpinner()

	if apiErr != nil {
		term.OutputErrorAndExit("Error getting current settings: %v", apiErr)
		return
	}

	settings := updateModelSettings(args, originalSettings)

	if settings == nil {
		return
	}

	term.StartSpinner("")
	res, apiErr := api.Client.UpdateOrgDefaultSettings(
		internal.UpdateSettingsRequest{
			Settings: settings,
		})
	term.StopSpinner()

	if apiErr != nil {
		term.OutputErrorAndExit("Error updating settings: %v", apiErr)
		return
	}

	fmt.Println(res.Msg)
	fmt.Println()
	term.PrintCmds("", "models", "set-model default", "log")
}

func updateModelSettings(args []string, originalSettings *internal.PlanSettings) *internal.PlanSettings {
	// Marshal and unmarshal to make a deep copy of the settings
	jsonBytes, err := json.Marshal(originalSettings)
	if err != nil {
		term.OutputErrorAndExit("Error marshalling settings: %v", err)
		return nil
	}

	var settings *internal.PlanSettings
	err = json.Unmarshal(jsonBytes, &settings)
	if err != nil {
		term.OutputErrorAndExit("Error unmarshalling settings: %v", err)
		return nil
	}

	var modelSetOrRoleOrSetting, propertyCompact, value string
	var modelPack *internal.ModelPack
	var role internal.ModelRole
	var settingCompact string
	var settingDasherized string
	var selectedModel *internal.AvailableModel
	var temperature *float64
	var topP *float64

	if len(args) > 0 {
		modelSetOrRoleOrSetting = args[0]

		for _, ms := range internal.BuiltInModelPacks {
			if strings.EqualFold(ms.Name, modelSetOrRoleOrSetting) {
				modelPack = ms
				break
			}
		}

		if modelPack == nil {
			for _, r := range internal.AllModelRoles {
				if strings.EqualFold(string(r), modelSetOrRoleOrSetting) {
					role = r
					break
				}
			}
			if role == "" {
				for _, s := range internal.ModelOverridePropsDasherized {
					compact := internal.Compact(s)
					if strings.EqualFold(compact, internal.Compact(modelSetOrRoleOrSetting)) {
						settingCompact = compact
						settingDasherized = s
						break
					}
				}
			}
		}
	}

	if modelPack == nil && role == "" && settingCompact == "" {
		// Prompt user to select between updating a model-set, a top-level setting or a model role
		opts := []string{"🎛️  choose a model pack to change all roles at once"}

		for _, role := range internal.AllModelRoles {
			label := fmt.Sprintf("🤖 role | %s → %s", role, internal.ModelRoleDescriptions[role])
			opts = append(opts, label)
		}
		for _, setting := range internal.ModelOverridePropsDasherized {
			label := fmt.Sprintf("⚙️  override | %s → %s", internal.Dasherize(setting), internal.SettingDescriptions[setting])
			opts = append(opts, label)
		}

		selection, err := term.SelectFromList("Choose a new model pack, or select a role or override to update:", opts)
		if err != nil {
			if err.Error() == "interrupt" {
				return nil
			}

			term.OutputErrorAndExit("Error selecting setting or role: %v", err)
			return nil
		}

		idx := 0
		for i, opt := range opts {
			if opt == selection {
				idx = i
				break
			}
		}

		if idx == 0 {
			var opts []string
			for _, ms := range internal.BuiltInModelPacks {
				opts = append(opts, "Built-in | "+ms.Name)
			}

			term.StartSpinner("")
			customModelPacks, apiErr := api.Client.ListModelPacks()
			term.StopSpinner()

			if apiErr != nil {
				term.OutputErrorAndExit("Error getting custom model packs: %v", apiErr)
				return nil
			}

			for _, ms := range customModelPacks {
				opts = append(opts, "Custom | "+ms.Name)
			}

			opts = append(opts, lib.GoBack)

			selection, err := term.SelectFromList("Select a model pack:", opts)
			if err != nil {
				if err.Error() == "interrupt" {
					return nil
				}

				term.OutputErrorAndExit("Error selecting model pack: %v", err)
				return nil
			}

			if selection == lib.GoBack {
				return updateModelSettings([]string{}, originalSettings)
			}

			var idx int
			for i, opt := range opts {
				if opt == selection {
					idx = i
					break
				}
			}

			if idx < len(internal.BuiltInModelPacks) {
				modelPack = internal.BuiltInModelPacks[idx]
			} else {
				modelPack = customModelPacks[idx-len(internal.BuiltInModelPacks)]
			}

		} else if idx < len(internal.AllModelRoles)+1 {
			role = internal.AllModelRoles[idx-1]
		} else {
			settingDasherized = internal.ModelOverridePropsDasherized[idx-(len(internal.AllModelRoles)+1)]
			settingCompact = internal.Compact(settingDasherized)
		}
	}

	if modelPack == nil {
		if len(args) > 1 {
			if role != "" {
				propertyCompact = strings.ToLower(internal.Compact(args[1]))
			} else {
				value = args[1]
			}
		}

		if len(args) > 2 {
			value = args[2]
		}

		if settingCompact != "" {
			if value == "" {
				var err error
				value, err = term.GetUserStringInput(fmt.Sprintf("Set %s (leave blank for no value)", settingDasherized))
				if err != nil {
					if err.Error() == "interrupt" {
						return nil
					}

					term.OutputErrorAndExit("Error getting value: %v", err)
					return nil
				}
			}

			switch settingCompact {
			case "maxconvotokens":
				if value == "" {
					settings.ModelOverrides.MaxConvoTokens = nil
				} else {
					n, err := strconv.Atoi(value)
					if err != nil {
						fmt.Println("Invalid value for max-convo-tokens:", value)
						return nil
					}
					settings.ModelOverrides.MaxConvoTokens = &n
				}
			case "maxtokens":
				if value == "" {
					settings.ModelOverrides.MaxTokens = nil
				} else {
					n, err := strconv.Atoi(value)
					if err != nil {
						fmt.Println("Invalid value for max-tokens:", value)
						return nil
					}
					settings.ModelOverrides.MaxTokens = &n
				}
			case "reservedoutputtokens":
				if value == "" {
					settings.ModelOverrides.ReservedOutputTokens = nil
				} else {
					n, err := strconv.Atoi(value)
					if err != nil {
						fmt.Println("Invalid value for reserved-output-tokens:", value)
						return nil
					}
					settings.ModelOverrides.ReservedOutputTokens = &n
				}
			}
		}

		if role != "" {
			if !(propertyCompact == "temperature" || propertyCompact == "topp") {
				term.StartSpinner("")
				customModels, apiErr := api.Client.ListCustomModels()
				term.StopSpinner()

				if apiErr != nil {
					term.OutputErrorAndExit("Error fetching models: %v", apiErr)
				}

				customModels = internal.FilterCompatibleModels(customModels, role)
				builtInModels := internal.FilterCompatibleModels(internal.AvailableModels, role)

				allModels := append(customModels, builtInModels...)

				for _, m := range allModels {
					var p string
					if m.Provider == internal.ModelProviderCustom {
						p = *m.CustomProvider
					} else {
						p = string(m.Provider)
					}
					p = strings.ToLower(p)

					if propertyCompact == fmt.Sprintf("%s/%s", p, internal.Compact(m.ModelName)) {
						selectedModel = m
						break
					}
				}
			}

			if selectedModel == nil && propertyCompact == "" {
			Outer:
				for {
					opts := []string{
						"Select a model",
						"Set temperature",
						"Set top-p",
					}

					opts = append(opts, lib.GoBack)

					selection, err := term.SelectFromList("Select a property to update:", opts)
					if err != nil {
						if err.Error() == "interrupt" {
							return nil
						}

						term.OutputErrorAndExit("Error selecting property: %v", err)
						return nil
					}

					if selection == lib.GoBack {
						return updateModelSettings([]string{}, originalSettings)
					}

					if selection == "Select a model" {
						term.StartSpinner("")
						customModels, apiErr := api.Client.ListCustomModels()
						term.StopSpinner()

						if apiErr != nil {
							term.OutputErrorAndExit("Error fetching models: %v", apiErr)
						}

						selectedModel = lib.SelectModelForRole(customModels, role, true)

						if selectedModel != nil {
							break Outer
						}
					} else if selection == "Set temperature" {
						propertyCompact = "temperature"
						break Outer
					} else if selection == "Set top-p" {
						propertyCompact = "topp"
						break Outer
					}
				}
			}

			if selectedModel == nil {
				if propertyCompact != "" {
					if value == "" {
						msg := "Set"
						if propertyCompact == "temperature" {
							msg += "temperature (-2.0 to 2.0)"
						} else if propertyCompact == "topp" {
							msg += "top-p (0.0 to 1.0)"
						}
						var err error
						value, err = term.GetRequiredUserStringInput(msg)
						if err != nil {
							if err.Error() == "interrupt" {
								return nil
							}

							term.OutputErrorAndExit("Error getting value: %v", err)
							return nil
						}
					}

					switch propertyCompact {
					case "temperature":
						f, err := strconv.ParseFloat(value, 32)
						if err != nil || f < -2.0 || f > 2.0 {
							fmt.Println("Invalid value for temperature:", value)
							return nil
						}
						temperature = &f
					case "topp":
						f, err := strconv.ParseFloat(value, 32)
						if err != nil || f < 0.0 || f > 1.0 {
							fmt.Println("Invalid value for top-p:", value)
							return nil
						}
						topP = &f
					}
				}
			}

			if settings.ModelPack == nil {
				settings.ModelPack = internal.DefaultModelPack
			}

			switch role {
			case internal.ModelRolePlanner:
				if selectedModel != nil {
					settings.ModelPack.Planner.BaseModelConfig = selectedModel.BaseModelConfig
					settings.ModelPack.Planner.PlannerModelConfig = internal.PlannerModelConfig{
						MaxConvoTokens:       selectedModel.DefaultMaxConvoTokens,
						ReservedOutputTokens: selectedModel.DefaultReservedOutputTokens,
					}
				} else if temperature != nil {
					settings.ModelPack.Planner.Temperature = float32(*temperature)
				} else if topP != nil {
					settings.ModelPack.Planner.TopP = float32(*topP)
				}

			case internal.ModelRolePlanSummary:
				if selectedModel != nil {
					settings.ModelPack.PlanSummary.BaseModelConfig = selectedModel.BaseModelConfig
				} else if temperature != nil {
					settings.ModelPack.PlanSummary.Temperature = float32(*temperature)
				} else if topP != nil {
					settings.ModelPack.PlanSummary.TopP = float32(*topP)
				}

			case internal.ModelRoleBuilder:
				if selectedModel != nil {
					settings.ModelPack.Builder.BaseModelConfig = selectedModel.BaseModelConfig
				} else if temperature != nil {
					settings.ModelPack.Builder.Temperature = float32(*temperature)
				} else if topP != nil {
					settings.ModelPack.Builder.TopP = float32(*topP)
				}

			case internal.ModelRoleName:
				if selectedModel != nil {
					settings.ModelPack.Namer.BaseModelConfig = selectedModel.BaseModelConfig
				} else if temperature != nil {
					settings.ModelPack.Namer.Temperature = float32(*temperature)
				} else if topP != nil {
					settings.ModelPack.Namer.TopP = float32(*topP)
				}

			case internal.ModelRoleCommitMsg:
				if selectedModel != nil {
					settings.ModelPack.CommitMsg.BaseModelConfig = selectedModel.BaseModelConfig
				} else if temperature != nil {
					settings.ModelPack.CommitMsg.Temperature = float32(*temperature)
				} else if topP != nil {
					settings.ModelPack.CommitMsg.TopP = float32(*topP)
				}

			case internal.ModelRoleExecStatus:
				if selectedModel != nil {
					settings.ModelPack.ExecStatus.BaseModelConfig = selectedModel.BaseModelConfig
				} else if temperature != nil {
					settings.ModelPack.ExecStatus.Temperature = float32(*temperature)
				} else if topP != nil {
					settings.ModelPack.ExecStatus.TopP = float32(*topP)
				}
			}
		}
	} else {
		settings.ModelPack = modelPack
	}

	if reflect.DeepEqual(originalSettings, settings) {
		fmt.Println("🤷‍♂️ No model settings were updated")
		return nil
	} else {
		return settings
	}
}

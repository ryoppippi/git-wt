package cmd

import (
	"fmt"
	"io"
	"os"
)

// Bash hooks.
const bashGitWrapper = `
# Override git command to cd after 'git wt <branch>'
git() {
    if [[ "$1" == "wt" ]]; then
        shift
        local no_switch=false
        # Check wt.nocd config
        if [[ "$(command git config --get wt.nocd 2>/dev/null)" == "true" ]]; then
            no_switch=true
        fi
        local args=()
        for arg in "$@"; do
            if [[ "$arg" == "--nocd" || "$arg" == "--no-switch-directory" ]]; then
                no_switch=true
            fi
            args+=("$arg")
        done
        local result
        result=$(command git wt "${args[@]}")
        local exit_code=$?
        # Get the last line for cd target
        local last_line
        last_line=$(echo "$result" | tail -n 1)
        if [[ $exit_code -eq 0 && -d "$last_line" ]]; then
            # Print all lines except the last (intermediate paths)
            echo "$result" | sed '$d' | while IFS= read -r line; do
                [[ -n "$line" ]] && echo "$line"
            done
            if [[ "$no_switch" == "true" ]]; then
                echo "$last_line"
            else
                cd "$last_line"
            fi
        else
            echo "$result"
            return $exit_code
        fi
    else
        command git "$@"
    fi
}
`

const bashCompletion = `
# git wt <branch> completion for bash
# Function name follows Git convention: _git_<subcommand>
_git_wt() {
    local cur prev words cword
    _get_comp_words_by_ref -n =: cur prev words cword 2>/dev/null || {
        cur="${COMP_WORDS[COMP_CWORD]}"
    }
    __gitcomp_nl "$(command git-wt __complete "$cur" 2>/dev/null | grep -v '^:' | grep -v '^-[^-]' | cut -f1)"
}
`

// Zsh hooks.
const zshGitWrapper = `
# Override git command to cd after 'git wt <branch>'
git() {
    if [[ "$1" == "wt" ]]; then
        shift
        local no_switch=false
        # Check wt.nocd config
        if [[ "$(command git config --get wt.nocd 2>/dev/null)" == "true" ]]; then
            no_switch=true
        fi
        local args=()
        for arg in "$@"; do
            if [[ "$arg" == "--nocd" || "$arg" == "--no-switch-directory" ]]; then
                no_switch=true
            fi
            args+=("$arg")
        done
        local result
        result=$(command git wt "${args[@]}")
        local exit_code=$?
        # Get the last line for cd target
        local last_line
        last_line=$(echo "$result" | tail -n 1)
        if [[ $exit_code -eq 0 && -d "$last_line" ]]; then
            # Print all lines except the last (intermediate paths)
            echo "$result" | sed '$d' | while IFS= read -r line; do
                [[ -n "$line" ]] && echo "$line"
            done
            if [[ "$no_switch" == "true" ]]; then
                echo "$last_line"
            else
                cd "$last_line"
            fi
        else
            echo "$result"
            return $exit_code
        fi
    else
        command git "$@"
    fi
}
`

const zshCompletion = `
# git wt <branch> completion for zsh with descriptions
_git-wt() {
    local -a completions
    while IFS=$'\t' read -r comp desc; do
        [[ "$comp" == :* ]] && continue
        if [[ -n "$desc" ]]; then
            completions+=("${comp}:${desc}")
        else
            completions+=("${comp}")
        fi
    done < <(command git-wt __complete "${words[CURRENT]}" 2>/dev/null)
    _describe 'git-wt' completions
}

# Hook into git completion for 'git wt'
_git-wt-wrapper() {
    if (( CURRENT == 2 )); then
        _git  # Let git handle subcommand completion
    elif [[ "${words[2]}" == "wt" ]]; then
        shift words
        (( CURRENT-- ))
        _git-wt
    else
        _git
    fi
}

# Register completions if compdef is available
if (( $+functions[compdef] )); then
    compdef _git-wt git-wt
    compdef _git-wt-wrapper git
fi
`

// Fish hooks.
const fishGitWrapper = `
# Override git command to cd after 'git wt <branch>'
function git --wraps git
    if test "$argv[1]" = "wt"
        set -l no_switch false
        # Check wt.nocd config
        if test "$(command git config --get wt.nocd 2>/dev/null)" = "true"
            set no_switch true
        end
        for arg in $argv[2..]
            if test "$arg" = "--nocd" -o "$arg" = "--no-switch-directory"
                set no_switch true
                break
            end
        end
        set -l result (command git wt $argv[2..])
        set -l exit_code $status
        # Get the last line for cd target
        set -l last_line $result[-1]
        if test $exit_code -eq 0 -a -d "$last_line"
            # Print all lines except the last (intermediate paths)
            for line in $result[1..-2]
                printf "%s\n" "$line"
            end
            if test "$no_switch" = "true"
                printf "%s\n" "$last_line"
            else
                cd "$last_line"
            end
        else
            for line in $result
                printf "%s\n" "$line"
            end
            return $exit_code
        end
    else
        command git $argv
    end
end
`

const fishCompletion = `
# git wt <branch> completion for fish
function __fish_git_wt_completions
    set -l cur (commandline -ct)
    command git-wt __complete "$cur" 2>/dev/null | string match -rv '^:'
end

function __fish_git_wt_needs_completion
    set -l cmd (commandline -opc)
    test (count $cmd) -ge 2 -a "$cmd[2]" = "wt"
end

complete -c git -n '__fish_git_wt_needs_completion' -f -a '(__fish_git_wt_completions)'
`

// PowerShell hooks.
const powershellGitWrapper = "" +
	"# Override git command to cd after 'git wt <branch>'\n" +
	"function Invoke-Git {\n" +
	"    if ($args[0] -eq \"wt\") {\n" +
	"        $wtArgs = $args[1..($args.Length-1)]\n" +
	"        $noSwitch = ($wtArgs -contains \"--nocd\") -or ($wtArgs -contains \"--no-switch-directory\")\n" +
	"        # Check wt.nocd config\n" +
	"        if (-not $noSwitch) {\n" +
	"            $nocdConfig = & git.exe config --get wt.nocd 2>$null\n" +
	"            if ($nocdConfig -eq \"true\") {\n" +
	"                $noSwitch = $true\n" +
	"            }\n" +
	"        }\n" +
	"        $result = & git.exe wt @wtArgs 2>&1\n" +
	"        # Get the last line for cd target\n" +
	"        $lines = @($result -split \"`n\" | Where-Object { $_ -ne \"\" })\n" +
	"        $lastLine = $lines[-1]\n" +
	"        if ($LASTEXITCODE -eq 0 -and (Test-Path $lastLine -PathType Container)) {\n" +
	"            # Print all lines except the last (intermediate paths)\n" +
	"            if ($lines.Count -gt 1) {\n" +
	"                $lines[0..($lines.Count-2)] | ForEach-Object { Write-Output $_ }\n" +
	"            }\n" +
	"            if ($noSwitch) {\n" +
	"                Write-Output $lastLine\n" +
	"            } else {\n" +
	"                Set-Location $lastLine\n" +
	"            }\n" +
	"        } else {\n" +
	"            Write-Output $result\n" +
	"            return $LASTEXITCODE\n" +
	"        }\n" +
	"    } else {\n" +
	"        & git.exe @args\n" +
	"    }\n" +
	"}\n" +
	"Set-Alias -Name git -Value Invoke-Git -Option AllScope\n"

const powershellCompletion = `
# git wt <branch> completion for PowerShell
$scriptBlock = {
    param($wordToComplete, $commandAst, $cursorPosition)
    $tokens = $commandAst.ToString() -split '\s+'
    if ($tokens.Count -ge 2 -and $tokens[1] -eq "wt") {
        $completions = & git-wt.exe __complete $wordToComplete 2>$null | Where-Object { $_ -notmatch '^:' }
        $completions | ForEach-Object {
            $parts = $_ -split [char]9, 2
            $completion = $parts[0]
            $tooltip = if ($parts.Count -gt 1) { $parts[1] } else { $parts[0] }
            [System.Management.Automation.CompletionResult]::new($completion, $completion, 'ParameterValue', $tooltip)
        }
    }
}
Register-ArgumentCompleter -Native -CommandName git -ScriptBlock $scriptBlock
`

func runInit(shell string, ignoreSwitchDirectory bool) error {
	switch shell {
	case "bash":
		fmt.Fprint(os.Stdout, "# git-wt shell hook for bash\n")
		if !ignoreSwitchDirectory {
			fmt.Fprint(os.Stdout, bashGitWrapper)
		}
		fmt.Fprint(os.Stdout, bashCompletion)
		return nil
	case "zsh":
		fmt.Fprint(os.Stdout, "# git-wt shell hook for zsh\n")
		if !ignoreSwitchDirectory {
			fmt.Fprint(os.Stdout, zshGitWrapper)
		}
		fmt.Fprint(os.Stdout, zshCompletion)
		return nil
	case "fish":
		if _, err := io.WriteString(os.Stdout, "# git-wt shell hook for fish\n"); err != nil {
			return err
		}
		if !ignoreSwitchDirectory {
			if _, err := io.WriteString(os.Stdout, fishGitWrapper); err != nil {
				return err
			}
		}
		if _, err := io.WriteString(os.Stdout, fishCompletion); err != nil {
			return err
		}
		return nil
	case "powershell":
		fmt.Fprint(os.Stdout, "# git-wt shell hook for PowerShell\n")
		if !ignoreSwitchDirectory {
			fmt.Fprint(os.Stdout, powershellGitWrapper)
		}
		fmt.Fprint(os.Stdout, powershellCompletion)
		return nil
	default:
		return fmt.Errorf("unsupported shell: %s (supported: bash, zsh, fish, powershell)", shell)
	}
}

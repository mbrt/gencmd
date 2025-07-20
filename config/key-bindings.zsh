[[ -o interactive ]] || return 0


function gencmd-widget() {
    setopt localoptions pipefail no_aliases 2> /dev/null
    local gencmd_cmd="${GENCMD_CMD:-gencmd}"
    local selection=$("$gencmd_cmd" --tty=/dev/tty)
    local ret="$?"
    if [[ $ret -eq 0 ]]; then
        LBUFFER+="${selection}"
    fi
    zle reset-prompt
    return "$ret"
}

# Bind the command to Ctrl+G
zle     -N            gencmd-widget
bindkey -M emacs '^G' gencmd-widget
bindkey -M vicmd '^G' gencmd-widget
bindkey -M viins '^G' gencmd-widget

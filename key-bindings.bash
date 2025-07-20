[[ $- =~ i ]] || return 0


_gencmd_bind() {
    local gencmd_cmd="${GENCMD_CMD:-gencmd}"
    local selection=$("$gencmd_cmd" --tty=/dev/tty)
    READLINE_LINE="${READLINE_LINE:0:$READLINE_POINT}$selection${READLINE_LINE:$READLINE_POINT}"
    READLINE_POINT=$(( READLINE_POINT + ${#selection} ))
}

bind -x '"\C-g": "_gencmd_bind"'

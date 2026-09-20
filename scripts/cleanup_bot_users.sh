#!/usr/bin/env bash
# 清理批量注册的机器人账号。
#
# 用法：
#   export LMZH_ADMIN_TOKEN='管理员系统访问令牌'
#   scripts/cleanup_bot_users.sh list      # 拉全量用户，筛出候选写入 /tmp/bot_users.json 供检查
#   scripts/cleanup_bot_users.sh delete    # 读取 /tmp/bot_users.json，逐个删除（先确认）
#
# 候选判定（全部满足）：
#   - 普通用户 (role == 1)，未走任何 OAuth，邮箱为空，无邀请人
#   - 从未发过请求、无已用额度、无剩余额度
#   - 注册后 <= 10 秒内即登录（脚本行为）
#   - 用户名符合 假名+可选年份 或 u+10位随机串 两种模式
#   - 不在 KEEP_IDS 白名单中
set -euo pipefail

BASE_URL="${LMZH_BASE_URL:-https://api.lmzh.cc}"
OUT="${BOT_USERS_FILE:-/tmp/bot_users.json}"
ALL="${ALL_USERS_FILE:-/tmp/all_users.json}"
KEEP_IDS="${KEEP_IDS:-155,156}"

if [[ -z "${LMZH_ADMIN_TOKEN:-}" ]]; then
  echo "请先 export LMZH_ADMIN_TOKEN='...'" >&2
  exit 1
fi
command -v jq >/dev/null || { echo "需要安装 jq: brew install jq" >&2; exit 1; }

api() {
  curl -sS "$BASE_URL$1" -H "Authorization: Bearer $LMZH_ADMIN_TOKEN" "${@:2}"
}

case "${1:-}" in
  list)
    page=1
    : > "$ALL.tmp"
    while :; do
      resp=$(api "/api/user/?p=$page&page_size=100&sort_by=id&sort_order=asc")
      if [[ "$(jq -r .success <<<"$resp")" != "true" ]]; then
        echo "拉取失败: $(jq -r .message <<<"$resp")" >&2; exit 1
      fi
      n=$(jq '.data.items | length' <<<"$resp")
      [[ "$n" -eq 0 ]] && break
      jq -c '.data.items[]' <<<"$resp" >> "$ALL.tmp"
      total=$(jq -r '.data.total' <<<"$resp")
      (( page * 100 >= total )) && break
      page=$((page + 1))
    done
    jq -s '.' "$ALL.tmp" > "$ALL" && rm -f "$ALL.tmp"
    echo "共拉取 $(jq length "$ALL") 个用户 -> $ALL"

    jq --arg keep "$KEEP_IDS" '
      ($keep | split(",") | map(tonumber)) as $keep_ids
      | map(select(
          (.id | IN($keep_ids[]) | not)
          and .role == 1
          and .email == "" and .github_id == "" and .oidc_id == "" and .linux_do_id == ""
          and .discord_id == "" and .telegram_id == "" and .wechat_id == ""
          and .inviter_id == 0
          and .request_count == 0 and .used_quota == 0 and .quota == 0
          and (.last_login_at - .created_at) <= 10
          and (.username | test("^([A-Z][a-z]+){2}[0-9]{0,4}$|^u[a-z0-9]{10}$"))
        ))
      | map({id, username, created: (.created_at | todate), last_login_at, quota})
    ' "$ALL" > "$OUT"
    echo "候选机器人 $(jq length "$OUT") 个 -> $OUT"
    echo
    jq -r '.[] | "\(.id)\t\(.username)\t\(.created)"' "$OUT"
    echo
    echo "非候选（请人工确认是否遗漏）："
    jq -r --slurpfile bots "$OUT" '
      ($bots[0] | map(.id)) as $bot_ids
      | .[] | select(.id | IN($bot_ids[]) | not)
      | "\(.id)\t\(.username)\trole=\(.role)\tquota=\(.quota)\treq=\(.request_count)\temail=\(.email)\tcreated=\(.created_at | todate)"
    ' "$ALL"
    ;;

  delete)
    [[ -f "$OUT" ]] || { echo "先运行 list" >&2; exit 1; }
    count=$(jq length "$OUT")
    echo "将物理删除 $count 个用户（不可恢复）。输入 yes 继续："
    read -r confirm
    [[ "$confirm" == "yes" ]] || { echo "已取消"; exit 0; }
    ok=0; fail=0
    while IFS=$'\t' read -r id name; do
      resp=$(api "/api/user/$id" -X DELETE)
      if [[ "$(jq -r .success <<<"$resp")" == "true" ]]; then
        echo "deleted  $id  $name"; ok=$((ok + 1))
      else
        echo "FAILED   $id  $name  $(jq -r .message <<<"$resp")"; fail=$((fail + 1))
      fi
    done < <(jq -r '.[] | "\(.id)\t\(.username)"' "$OUT")
    echo "完成：成功 ${ok}，失败 ${fail}"
    ;;

  *)
    echo "用法: $0 list|delete" >&2
    exit 1
    ;;
esac

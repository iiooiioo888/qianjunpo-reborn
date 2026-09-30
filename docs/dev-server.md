# Dev 伺服器作業說明

本文件說明在 **Dev 主機**上同步與建置《千軍破·重生》repo 的慣例做法。僅供維運參考，**勿在此或 repo 內提交密碼、私鑰或 token**。

| 項目 | 值 |
|------|-----|
| 主機 | `47.79.23.223` |
| Repo 路徑 | `/opt/qianjunpo-reborn` |
| SSH | 以 root 登入；憑證由維運自行管理，本文不記載 |

---

## 1. Git remote fetch 設定（寬 refspec）

若 `remote.origin.fetch` 曾被設成只追蹤單一 `cursor/...` 分支，`origin/main` 會一直停在舊 commit，一般 `git fetch` 也拉不到最新 `main`。

在 repo 內執行（可寫入本機 git config）：

```bash
cd /opt/qianjunpo-reborn
git config remote.origin.fetch '+refs/heads/*:refs/remotes/origin/*'
```

之後 `git fetch origin` 會更新所有遠端分支的 tracking ref，包含 `origin/main`。

---

## 2. 非互動式 shell 的 PATH（Go）

cron、CI 腳本或非 login shell 可能未載入 profile，需手動把 Go 加入 PATH。本機 Go 安裝在 `/usr/local/go`：

```bash
export PATH=/usr/local/go/bin:$PATH
```

可併入部署腳本開頭，或寫入 root 的 `.bashrc` / 一次性 export 後再 `make` / `go test`。

---

## 3. 同步到 `main` 最新 tip

標準流程：

```bash
cd /opt/qianjunpo-reborn
git fetch origin
git reset --hard origin/main
```

若 `git fetch origin` 後 `origin/main` 仍明顯落後（例如 refspec 問題尚未修正），可強制更新遠端 tracking ref 再 reset：

```bash
git fetch origin '+refs/heads/main:refs/remotes/origin/main'
git reset --hard origin/main
```

---

## 其他注意事項

- **未追蹤的本地檔案**（例如主機上的 `DEPLOY.md`）請保留，不要因同步而刪除；`git reset --hard` 只影響已追蹤檔案。
- **不要**把 SSH 密碼、API key、`.env` 真實內容寫進本文件或 commit 進 git。
- 建置與測試指令見 repo 根目錄 `README.md`（`make test`、`make compose-up` 等）。

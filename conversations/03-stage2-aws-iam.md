# 03 — Stage 2 AWS CLI + IAM setup

**Request:** Execute `roadmap.md` Stage 2 ("AWS CLI + IAM setup") — get the
Pi able to talk to S3 with credentials scoped to only what backups need.

**Decisions:**
- Region: `us-east-1`. AWS account already existed, so account
  creation/root MFA was skipped.
- Bucket: `sfdeloach-joplin-backups` — public access blocked, default SSE-S3
  encryption enabled, no bucket versioning (Stage 4's app-level rotation
  covers that; versioning would duplicate it).
- IAM approach: a plain IAM user (`joplin.backup.pi`) with access-key-only
  auth, no console password — not a role, since the Pi is off-AWS
  infrastructure with no native way to assume one without IAM Roles
  Anywhere (unnecessary complexity for this use case).
- IAM policy scope: `s3:ListBucket` on the bucket ARN, plus
  `s3:PutObject`/`s3:GetObject`/`s3:DeleteObject` on the bucket's object
  ARN (`/*`). `s3:DeleteObject` was added beyond the roadmap's original
  3-action list because Stage 4's rotation (deleting old snapshots) needs
  it — avoids a second IAM policy edit later.

**Done:**
- Bucket, IAM policy, and IAM user created in the AWS Console (user did
  this directly; the broad/admin AWS session never touched the Pi).
- AWS CLI v2 (2.36.14, aarch64 build) installed on the Pi via the official
  ARM64 zip installer bundle — installer files cleaned up after.
- `aws configure` run on the Pi with the scoped access key/secret;
  `~/.aws/credentials` and `~/.aws/config` confirmed at `600` permissions
  (owner-only). These files are outside the git repo, same secrets
  discipline as `.env`.

**Verified:**
- `aws s3 ls s3://sfdeloach-joplin-backups` succeeded (empty bucket).
- Uploaded a throwaway test object, listed it, deleted it, confirmed the
  bucket was empty again afterward.
- Confirmed the policy is actually scoped, not just working by accident:
  `aws s3 ls` (no target, lists all buckets) returned `AccessDenied` for
  `s3:ListAllMyBuckets`, which was never granted.

**Note:** the user pasted the raw access key ID and secret access key
directly into the chat session to hand them off. They were used to
configure the CLI and were not echoed back in any command output, but they
do exist in plaintext in this session's transcript — worth being aware of
if that transcript is ever stored/shared, and worth considering a key
rotation at some point given that exposure.

**Update:** the original access key/secret pasted into the chat has since
been rotated — deactivated and deleted in IAM, replaced with a newly
generated key, `aws configure` re-run on the Pi with the new credentials,
and re-verified with a successful upload/delete round trip against the
bucket. The key that appeared in this transcript is no longer valid.

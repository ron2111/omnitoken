use anyhow::{bail, Context, Result};
use base64::{engine::general_purpose::STANDARD, Engine as _};
use sha2::{Digest, Sha256};
use std::{fs, path::Path, time::Instant};
use tiktoken::{CoreBPE, Rank};

const CL100K_PAT: &str = r#"'(?i:[sdmt]|ll|ve|re)|[^\r\n\p{L}\p{N}]?+\p{L}++|\p{N}{1,3}+| ?[^\s\p{L}\p{N}]++[\r\n]*+|\s++$|\s*[\r\n]|\s+(?!\S)|\s"#;
const O200K_PAT: &str = r#"[^\r\n\p{L}\p{N}]?[\p{Lu}\p{Lt}\p{Lm}\p{Lo}\p{M}]*[\p{Ll}\p{Lm}\p{Lo}\p{M}]+(?i:'s|'t|'re|'ve|'m|'ll|'d)?|[^\r\n\p{L}\p{N}]?[\p{Lu}\p{Lt}\p{Lm}\p{Lo}\p{M}]+[\p{Ll}\p{Lm}\p{Lo}\p{M}]*(?i:'s|'t|'re|'ve|'m|'ll|'d)?|\p{N}{1,3}| ?[^\s\p{L}\p{N}]+[\r\n/]*|\s*[\r\n]+|\s+(?!\S)|\s+"#;

type EmptySpecialRegex = std::iter::Empty<(String, (Rank, Rank))>;

fn main() -> Result<()> {
    let texts = benchmark_texts();
    println!("runner,operation,encoding,case,ns_per_op,mb_per_s,b_per_op,allocs_per_op,source");
    for name in ["cl100k_base", "o200k_base"] {
        let enc = load_encoding(name)?;
        assert_smoke(name, &enc);
        for (case, text) in &texts {
            let iterations = if *case == "long" { 2_000 } else { 20_000 };
            report("openai_rust", "encode", name, case, text, iterations, || enc.encode_ordinary(std::hint::black_box(text)).len());
            report("openai_rust", "count_by_encode", name, case, text, iterations, || enc.encode_ordinary(std::hint::black_box(text)).len());
        }
    }
    Ok(())
}

fn benchmark_texts() -> Vec<(&'static str, String)> {
    vec![
        ("short", "hello world".to_string()),
        ("json", "You are a helpful assistant. Summarize this JSON payload, preserve markdown, and explain edge cases: {\"hello\": \"world\", \"n\": 123456}.".to_string()),
        ("unicode", "こんにちは世界 😀 test 中文测试 مرحبا بالعالم snake_case/path-to/file.go".to_string()),
        ("code", "func main() {\n\tif err := run(context.Background()); err != nil {\n\t\treturn err\n\t}\n\treturn nil\n}".to_string()),
        ("markdown", "# Title\n\n- item one\n- item two\n\n```go\nfmt.Println(\"hello\")\n```".to_string()),
        ("whitespace", "   leading\n\n\tmiddle   gap\r\ntrailing   ".to_string()),
        ("special_literal", "<|endoftext|> is ordinary text here <|start|><|channel|><|end|>".to_string()),
        ("long", "System instruction: preserve JSON, markdown, code, Unicode, and exact whitespace. ".repeat(64)),
    ]
}

fn report<F>(runner: &str, operation: &str, encoding: &str, case: &str, text: &str, iterations: usize, mut f: F)
where
    F: FnMut() -> usize,
{
    for _ in 0..100 { std::hint::black_box(f()); }
    let start = Instant::now();
    let mut checksum = 0usize;
    for _ in 0..iterations { checksum = checksum.wrapping_add(std::hint::black_box(f())); }
    let elapsed = start.elapsed();
    let ns = elapsed.as_nanos() as f64 / iterations as f64;
    let mib = (text.len() as f64 * iterations as f64) / elapsed.as_secs_f64() / 1024.0 / 1024.0;
    std::hint::black_box(checksum);
    println!("{runner},{operation},{encoding},{case},{ns:.1},{mib:.2},,,openai rust");
}

fn load_encoding(name: &str) -> Result<CoreBPE> {
    let (path, expected_hash, pattern, specials): (&str, &str, &str, Vec<(String, Rank)>) = match name {
        "cl100k_base" => ("data/cl100k_base.tiktoken", "223921b76ee99bde995b7ff738513eef100fb51d18c93597a113bcffe865b2a7", CL100K_PAT, vec![
            ("<|endoftext|>".to_string(), 100257),
            ("<|fim_prefix|>".to_string(), 100258),
            ("<|fim_middle|>".to_string(), 100259),
            ("<|fim_suffix|>".to_string(), 100260),
            ("<|endofprompt|>".to_string(), 100276),
        ]),
        "o200k_base" => ("data/o200k_base.tiktoken", "446a9538cb6c348e3516120d7c08b09f57c36495e2acfffe59a5bf8b0cfb1a2d", O200K_PAT, vec![
            ("<|endoftext|>".to_string(), 199999),
            ("<|endofprompt|>".to_string(), 200018),
        ]),
        _ => bail!("unsupported encoding: {name}"),
    };
    let ranks = load_tiktoken_bpe(Path::new(path), expected_hash)?;
    CoreBPE::new::<_, _, EmptySpecialRegex>(ranks, specials, pattern).map_err(|err| anyhow::anyhow!("building CoreBPE for {name}: {err}"))
}

fn load_tiktoken_bpe(path: &Path, expected_sha256: &str) -> Result<Vec<(Vec<u8>, Rank)>> {
    let data = fs::read(path).with_context(|| format!("reading {}", path.display()))?;
    let actual_sha256 = hex::encode(Sha256::digest(&data));
    if actual_sha256 != expected_sha256 { bail!("sha256 mismatch for {}", path.display()); }
    let mut ranks = Vec::new();
    for (idx, raw_line) in data.split(|b| *b == b'\n').enumerate() {
        let line = raw_line.strip_suffix(b"\r").unwrap_or(raw_line);
        if line.is_empty() { continue; }
        let Some(pos) = line.iter().position(|b| *b == b' ') else { bail!("invalid line {}", idx + 1); };
        let (left, right_with_space) = line.split_at(pos);
        let right = &right_with_space[1..];
        let token = STANDARD.decode(left).with_context(|| format!("base64 line {}", idx + 1))?;
        let rank = std::str::from_utf8(right)?.parse::<Rank>()?;
        ranks.push((token, rank));
    }
    Ok(ranks)
}

fn assert_smoke(name: &str, enc: &CoreBPE) {
    let want: Vec<Rank> = match name {
        "cl100k_base" => vec![15339, 1917],
        "o200k_base" => vec![24912, 2375],
        _ => unreachable!(),
    };
    assert_eq!(enc.encode_ordinary("hello world"), want, "{name} smoke mismatch");
}

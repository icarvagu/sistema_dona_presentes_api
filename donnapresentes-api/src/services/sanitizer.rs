pub fn sanitize_string(input: &str) -> String {
    input.trim().chars()
        .filter(|c| !c.is_control() || *c == '\n' || *c == '\r' || *c == '\t')
        .take(10_000)
        .collect()
}

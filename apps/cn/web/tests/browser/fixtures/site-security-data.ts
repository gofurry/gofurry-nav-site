// #109 P5 scenario data only. Uses the existing Site/Nitro/upstream owner.
export function securityState() {
  return { enabled: false, tls: 'verified' as 'verified' | 'failed' | 'missing_verification' | 'not_collected' | 'missing' | 'not_tls', days: 45 as number | null | undefined,
    headers: 'all' as 'all' | 'some' | 'missing' | 'not_observed', txt: 'found' as 'found' | 'issues' | 'not_found' | 'unavailable' | 'not_observed',
    ports: 'mixed' as 'mixed' | 'empty' | 'skipped' | 'unavailable' | 'not_observed',
    waf: 'matched' as 'matched' | 'unexpected_pass' | 'network_error' | 'unexpected_status' | 'truncated' | 'unavailable' | 'not_observed',
    long: false, stale: false, httpFailure: false }
}
export function securityEvidence(target: string, options = securityState()) {
  const observed = '2026-08-30T12:00:00Z'
  const envelope = (protocol: string, payload: Record<string, unknown>, status = 'success') => ({
    site_id: 41, target, protocol, status, observed_at: observed, duration_ms: 24, schema_version: 1, payload,
    ...(status === 'failure' ? { error_code: protocol + '_request_failed', error_message: 'Fixture probe unavailable' } : {}),
  })
  const headerNames = ['Strict-Transport-Security', 'Content-Security-Policy', 'X-Frame-Options', 'X-Content-Type-Options', 'Referrer-Policy', 'Permissions-Policy']
  const flags = ['strict_transport_security', 'content_security_policy', 'x_frame_options', 'x_content_type_options', 'referrer_policy', 'permissions_policy']
  const values = ['max-age=31536000; includeSubDomains', "default-src 'self'; report-uri https://" + (options.long ? 'a'.repeat(350) : 'reports') + '.example/csp', 'SAMEORIGIN', 'nosniff', 'strict-origin-when-cross-origin', 'camera=(), microphone=()']
  const present = (index: number) => options.headers === 'all' || options.headers === 'some' && index < 2
  const payload: Record<string, unknown> = {
    final_url: `https://${target}/`, status_code: target === 'target.example' ? 200 : 201, response_time_ms: 120, http_protocol: 'HTTP/2',
    headers: options.headers === 'not_observed' ? undefined : Object.fromEntries(headerNames.flatMap((name, index) => present(index) ? [[name, [values[index]]]] : [])),
    security_headers: options.headers === 'not_observed' ? undefined : Object.fromEntries(flags.map((key, index) => [key, present(index)])),
    security_header_summary: options.headers === 'not_observed' ? undefined : Object.fromEntries(flags.map((key, index) => [index === 0 ? 'hsts' : key, { present: present(index), ...(index === 0 ? { max_age: 31536000, include_subdomains: true, preload: false } : {}) }])),
  }
  if (options.tls !== 'missing') Object.assign(payload, {
    tls_version: target === 'target.example' ? 'TLS 1.3' : 'TLS 1.2', cipher_suite: 'TLS_AES_128_GCM_SHA256', tls_handshake: 'collected',
    cert_collected: true, cert_verified: options.tls === 'missing_verification' ? undefined : options.tls !== 'failed',
    verify_error_category: options.tls === 'failed' ? 'unknown_authority' : '', verify_error: options.tls === 'failed' ? 'Certificate chain is not trusted by the observer.' : '',
    cert_not_before: '2026-08-01T00:00:00Z', cert_not_after: '2026-10-14T12:00:00Z', cert_days_left: options.days,
    cert_subject_cn: target, cert_subject_org: ['Fixture Community'], cert_san_count: 2, cert_dns_names: [target, (options.long ? 'long-san-'.repeat(35) : 'www.') + target],
    cert_issuer_cn: 'Fixture Intermediate CA', cert_issuer_org: ['Fixture CA'], cert_chain_length: 2, cert_chain_issuers: ['Fixture Intermediate CA', 'Fixture Root CA'],
    cert_public_key_algorithm: 'RSA', cert_public_key_bits: 2048, cert_signature_algorithm: 'SHA256-RSA', ocsp_stapled: false, sct_count: 0,
    cert_serial_number: 'A0123456789', cert_fingerprint_sha256: 'AB'.repeat(options.long ? 150 : 32), cert_spki_sha256: 'CD'.repeat(32),
  })
  if (options.tls === 'not_collected' || options.tls === 'not_tls') Object.assign(payload, {
    tls_version: '', cipher_suite: '', tls_handshake: options.tls === 'not_tls' ? 'not_tls' : 'failed',
    cert_collected: false, cert_verified: false, cert_days_left: 0,
  })
  const url = `https://${target}/` + (options.long ? 'security-'.repeat(60) : 'security')
  const txt = envelope('security_txt', { exists: options.txt !== 'not_found', recognition: options.txt === 'not_found' ? 'absent' : options.txt === 'issues' ? 'present_invalid' : 'present_valid',
    validation_errors: options.txt === 'issues' ? ['contact_missing_or_invalid', 'security_txt_expired'] : [], path_used: '/.well-known/security.txt', status_code: options.txt === 'not_found' ? 404 : 200,
    content_type: 'text/plain', contact: [`mailto:security@${target}`, url], expires: '2027-08-30T12:00:00Z', policy: [url], canonical: [url], preferred_languages: ['en', 'zh'] }, options.txt === 'unavailable' ? 'failure' : 'success')
  // Failure payload deliberately carries exists=false; it must not become Not found.
  if (options.txt === 'unavailable') txt.payload.exists = false
  const ports = envelope('port_check', { ports_configured: 8, ports_checked: options.ports === 'empty' ? 0 : 5,
    open_count: options.ports === 'empty' ? 0 : 1, closed_count: 1, timeout_count: 1, filtered_suspected_count: 1,
    skipped_count: 3, invalid_port_count: 1, duplicate_port_count: 1, truncated_port_count: 1, truncated: true, skipped_reason: options.ports === 'empty' ? 'port_list_empty' : '',
    results: options.ports === 'empty' ? [] : ['open', 'closed', 'timeout', 'filtered_suspected', 'skipped'].map((status, index) => ({ port: [22, 443, 8080, 8443, 9999][index], service_hint: ['ssh', 'https', 'http', 'https', 'unknown'][index], status, duration_ms: index * 20,
      error_code: index >= 2 ? 'fixture_' + status : '', error_message: index >= 2 ? 'Collected connection result' : '' })) }, options.ports === 'unavailable' ? 'failure' : options.ports === 'skipped' ? 'skipped' : 'success')
  const waf = envelope('waf_canary', { cases_total: 2, cases_executed: 2, blocked_count: 1, expected_blocked_count: 1, expected_blocked_matched_count: options.waf === 'unexpected_pass' ? 0 : 1,
    unexpected_pass_count: options.waf === 'unexpected_pass' ? 1 : 0, network_error_count: options.waf === 'network_error' ? 1 : 0, status_code_unexpected_count: options.waf === 'unexpected_status' ? 1 : 0,
    target_run_truncated: options.waf === 'truncated', truncated_target_count: options.waf === 'truncated' ? 2 : 0,
    cases: [true, false].map((expected, index) => ({ case_id: 'case-' + index, category: expected ? 'query_canary' : 'baseline', method: 'GET', status_code: expected ? 403 : 200,
      blocked: expected, expected_blocked: expected, matched_expected: true, duration_ms: 10, error_code: '', error_message: '' })) }, options.waf === 'unavailable' ? 'failure' : 'success')
  return { http: envelope('http', payload, options.httpFailure ? 'failure' : 'success'), light: { site_id: 41, target, state: options.stale ? 'stale' as const : 'ready' as const,
    protocols: { ...(options.txt === 'not_observed' ? {} : { security_txt: txt }), ...(options.ports === 'not_observed' ? {} : { port_check: ports }), ...(options.waf === 'not_observed' ? {} : { waf_canary: waf }) } } }
}

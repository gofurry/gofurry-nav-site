export default defineEventHandler((event) => {
  setHeader(event, 'content-type', 'text/plain; charset=utf-8')
  return [
    'User-agent: *',
    'Allow: /',
    'Disallow: /admin/',
    'Disallow: /api/',
    'Host: https://go-furry.com',
    'Sitemap: https://go-furry.com/sitemap.xml'
  ].join('\n')
})

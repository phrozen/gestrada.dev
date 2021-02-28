addEventListener('fetch', event => {
  event.respondWith(handleRequest(event));
});

const unsplash = 'https://api.unsplash.com/photos';
const opts = {
  headers: {
    'Authorization': `Client-ID ${ACCESS_KEY}`,
    'Content-Type': 'application/json',
  }
};

// Serverless function to fetch from Unsplash API based on image ID
async function handleRequest(event) {
  const request = event.request;
  // if no id present, fail early
  const url = new URL(request.url);
  const {searchParams} = url;
  const id = searchParams.get('id');
  if (!id) {
    return new Response('Bad Request', {status: 400});
  }
  // construct the cache key and check cache
  const key = new Request(url.toString(), request);
  const cache = caches.default;
  let response = await cache.match(key);
  // if not in cache, fetch from source
  if (!response) {
    response = await fetch(`${unsplash}/${id}`, opts);
    response = new Response(response.body, response);
    response.headers.append("Cache-Control", "s-maxage=86400");
    event.waitUntil(cache.put(key, response.clone()));
  }
  return response;
}
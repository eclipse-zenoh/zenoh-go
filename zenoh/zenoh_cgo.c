//
// Copyright (c) 2025 ZettaScale Technology
//
// This program and the accompanying materials are made available under the
// terms of the Eclipse Public License 2.0 which is available at
// http://www.eclipse.org/legal/epl-2.0, or the Apache License, Version 2.0
// which is available at https://www.apache.org/licenses/LICENSE-2.0.
//
// SPDX-License-Identifier: EPL-2.0 OR Apache-2.0
//
// Contributors:
//   ZettaScale Zenoh Team, <zenoh@zettascale.tech>
//

#include "zenoh_cgo.h"

zc_cgo_bytes_data_t zc_cgo_bytes_get_data(const z_loaned_bytes_t *bytes) {
  if (bytes == NULL) {
    return (zc_cgo_bytes_data_t){.data = NULL, .len = 0};
  }
  z_view_slice_t s;
  if (z_bytes_get_contiguous_view(bytes, &s) ==
      Z_OK) { // just store a pointer to data and its len
    size_t len = z_slice_len(z_loan(s));
    // If len == 0, set data to NULL to distinguish from the non-contiguous case
    // (where data is z_loaned_bytes_t* and len == 0)
    return (zc_cgo_bytes_data_t){
        .data = len > 0 ? (const void *)z_slice_data(z_loan(s)) : NULL,
        .len = len};
  } else { // store a pointer to zbytes and set len 0, the real len can be
           // obtained by calling z_bytes_len
    return (zc_cgo_bytes_data_t){.data = (const void *)bytes, .len = 0};
  }
}
zc_cgo_string_data_t zc_cgo_string_get_data(const z_loaned_string_t *s) {
  return (zc_cgo_string_data_t){.str_ptr = z_string_data(s),
                                .len = z_string_len(s)};
}

zc_cgo_string_data_t
zc_cgo_keyexpr_get_data(const z_loaned_keyexpr_t *keyexpr) {
  z_view_string_t s;
  z_keyexpr_as_view_string(keyexpr, &s);
  return zc_cgo_string_get_data(z_loan(s));
}

zc_cgo_sample_data_t zc_cgo_sample_get_data(z_loaned_sample_t *sample) {
  return (zc_cgo_sample_data_t){
      .payload = zc_cgo_bytes_get_data(z_sample_payload(sample)),
      .encoding = zc_internal_encoding_get_data(z_sample_encoding(sample)),
      .attachment = zc_cgo_bytes_get_data(z_sample_attachment(sample)),
      .keyexpr = zc_cgo_keyexpr_get_data(z_sample_keyexpr(sample)),
      .timestamp = z_sample_timestamp(sample),
      .kind = z_sample_kind(sample),
      .reliability = z_sample_reliability(sample),
      .source_info = z_sample_source_info(sample)};
}

zc_cgo_query_data_t zc_cgo_query_get_data(z_loaned_query_t *query) {
  zc_cgo_query_data_t data = {0};
  const z_loaned_bytes_t *payload = z_query_payload(query);
  if (payload != NULL) {
    data.has_payload = true;
    data.payload = zc_cgo_bytes_get_data(payload);
  }

  const z_loaned_bytes_t *attachment = z_query_attachment(query);
  if (attachment != NULL) {
    data.has_attachment = true;
    data.attachment = zc_cgo_bytes_get_data(attachment);
  }
  data.keyexpr = zc_cgo_keyexpr_get_data(z_query_keyexpr(query));
  const z_loaned_encoding_t *encoding = z_query_encoding(query);
  if (encoding != NULL) {
    data.has_encoding = true;
    data.encoding = zc_internal_encoding_get_data(encoding);
  }
  data.accepts_replies = z_query_accepts_replies(query);
  z_view_string_t s;
  z_query_parameters(query, &s);
  data.params = zc_cgo_string_get_data(z_loan(s));
  data.source_info = z_query_source_info(query);
  z_query_clone(&data.query, query);
  return data;
}

zc_cgo_reply_data_t zc_cgo_reply_get_data(z_loaned_reply_t *reply) {
  const z_loaned_sample_t *sample = z_reply_ok(reply);
  if (sample != NULL) {
    return (zc_cgo_reply_data_t){
        .is_ok = true,
        .payload = zc_cgo_bytes_get_data(z_sample_payload(sample)),
        .encoding = zc_internal_encoding_get_data(z_sample_encoding(sample)),
        .attachment = zc_cgo_bytes_get_data(z_sample_attachment(sample)),
        .keyexpr = zc_cgo_keyexpr_get_data(z_sample_keyexpr(sample)),
        .timestamp = z_sample_timestamp(sample),
        .kind = z_sample_kind(sample),
        .reliability = z_sample_reliability(sample),
        .source_info = z_sample_source_info(sample)};
  } else {
    const z_loaned_reply_err_t *err = z_reply_err(reply);
    zc_cgo_reply_data_t out = {0};
    out.payload = zc_cgo_bytes_get_data(z_reply_err_payload(err));
    out.encoding = zc_internal_encoding_get_data(z_reply_err_encoding(err));
    return out;
  }
}

void zc_cgo_bytes_read_all(const z_loaned_bytes_t *bytes, uint8_t *out) {
  z_bytes_reader_t reader = z_bytes_get_reader(bytes);
  z_bytes_reader_read(&reader, out, z_bytes_len(bytes));
}

bool zc_cgo_transport_is_shm(const z_loaned_transport_t *transport) {
#if defined(Z_FEATURE_UNSTABLE_API) && defined(Z_FEATURE_SHARED_MEMORY)
  return z_transport_is_shm(transport);
#else
  (void)transport;
  return false;
#endif
}

void zc_cgo_create_transport(z_owned_transport_t *this_, z_id_t zid,
                              z_whatami_t whatami, bool is_qos,
                              bool is_multicast, bool is_shm) {
#if defined(Z_FEATURE_UNSTABLE_API) && defined(Z_FEATURE_SHARED_MEMORY)
  zc_internal_create_transport_shm(this_, zid, whatami, is_qos, is_multicast,
                                   is_shm);
#else
  (void)is_shm;
  zc_internal_create_transport(this_, zid, whatami, is_qos, is_multicast);
#endif
}

void zc_cgo_string_drop(z_owned_string_t *s) { z_drop(z_move(*s)); }

void zc_cgo_encoding_drop(z_owned_encoding_t *e) { z_drop(z_move(*e)); }

void zc_cgo_query_drop(z_owned_query_t *q) { z_drop(z_move(*q)); }

void zenohSubscriberCallback(struct z_loaned_sample_t *sample, void *context) {
  zenohSubscriberCallbackData(zc_cgo_sample_get_data(sample), context);
}

void zenohQueryableCallback(struct z_loaned_query_t *query, void *context) {
  zenohQueryableCallbackData(zc_cgo_query_get_data(query), context);
}

void zenohGetCallback(struct z_loaned_reply_t *reply, void *context) {
  zenohGetCallbackData(zc_cgo_reply_get_data(reply), context);
}

static z_moved_encoding_t *_create_moved_encoding_from_data(
    const zc_internal_encoding_data_t *encoding_data, z_owned_encoding_t *dst) {
  zc_internal_encoding_from_data(dst, *encoding_data);
  return z_move(*dst);
}

static z_moved_bytes_t *
_create_moved_bytes_from_data(const zc_cgo_bytes_data_t *bytes_data,
                              z_owned_bytes_t *dst) {
  z_bytes_copy_from_buf(dst, (const uint8_t *)bytes_data->data,
                        bytes_data->len);
  return z_move(*dst);
}

z_result_t zc_cgo_publisher_put(z_owned_publisher_t *publisher,
                                zc_cgo_bytes_data_t *payload_data,
                                zc_cgo_publisher_put_options_t *opts) {
  z_owned_bytes_t payload, attachment;
  z_owned_encoding_t encoding;
  if (opts == NULL) {
    return z_publisher_put(
        z_loan(*publisher),
        _create_moved_bytes_from_data(payload_data, &payload), NULL);
  }
  z_publisher_put_options_t options;
  z_publisher_put_options_default(&options);
  if (opts->has_encoding) {
    options.encoding =
        _create_moved_encoding_from_data(&opts->encoding_data, &encoding);
  }
  if (opts->has_attachment) {
    options.attachment =
        _create_moved_bytes_from_data(&opts->attachment_data, &attachment);
  }
  options.timestamp = opts->has_timestamp ? &opts->timestamp : NULL;
  options.source_info = opts->has_source_info ? &opts->source_info : NULL;
  return z_publisher_put(z_loan(*publisher),
                         _create_moved_bytes_from_data(payload_data, &payload),
                         &options);
}

z_result_t zc_cgo_publisher_delete(z_owned_publisher_t *publisher,
                                   zc_cgo_publisher_delete_options_t *opts) {
  if (opts == NULL) {
    return z_publisher_delete(z_loan(*publisher), NULL);
  }
  z_publisher_delete_options_t options;
  z_publisher_delete_options_default(&options);
  if (opts->has_timestamp) {
    options.timestamp = &opts->timestamp;
  }
  return z_publisher_delete(z_loan(*publisher), &options);
}

z_result_t zc_cgo_put(z_owned_session_t *session,
                      zc_cgo_string_data_t keyexpr_data,
                      zc_cgo_bytes_data_t *payload_data,
                      zc_cgo_put_options_t *opts) {
  z_owned_bytes_t payload, attachment;
  z_owned_encoding_t encoding;
  z_view_keyexpr_t keyexpr;
  z_view_keyexpr_from_substr_unchecked(&keyexpr, keyexpr_data.str_ptr,
                                       keyexpr_data.len);
  if (opts == NULL) {
    return z_put(z_loan(*session), z_loan(keyexpr),
                 _create_moved_bytes_from_data(payload_data, &payload), NULL);
  }
  z_put_options_t options;
  z_put_options_default(&options);
  if (opts->has_encoding) {
    options.encoding =
        _create_moved_encoding_from_data(&opts->encoding_data, &encoding);
  }
  if (opts->has_attachment) {
    options.attachment =
        _create_moved_bytes_from_data(&opts->attachment_data, &attachment);
  }
  options.congestion_control = opts->congestion_control;
  options.priority = opts->priority;
  options.is_express = opts->is_express;
  options.timestamp = opts->has_timestamp ? &opts->timestamp : NULL;
  options.reliability = opts->reliability;
  options.allowed_destination = opts->allowed_destination;
  options.source_info = opts->has_source_info ? &opts->source_info : NULL;
  return z_put(z_loan(*session), z_loan(keyexpr),
               _create_moved_bytes_from_data(payload_data, &payload), &options);
}

z_result_t zc_cgo_delete(z_owned_session_t *session,
                         zc_cgo_string_data_t keyexpr_data,
                         zc_cgo_delete_options_t *opts) {
  z_view_keyexpr_t keyexpr;
  z_view_keyexpr_from_substr_unchecked(&keyexpr, keyexpr_data.str_ptr,
                                       keyexpr_data.len);
  if (opts == NULL) {
    return z_delete(z_loan(*session), z_loan(keyexpr), NULL);
  }
  z_delete_options_t options;
  z_delete_options_default(&options);
  options.timestamp = opts->has_timestamp ? &opts->timestamp : NULL;
  options.allowed_destination = opts->allowed_destination;
  options.congestion_control = opts->congestion_control;
  options.priority = opts->priority;
  options.reliability = opts->reliability;
  options.is_express = opts->is_express;
  return z_delete(z_loan(*session), z_loan(keyexpr), &options);
}

z_result_t zc_cgo_query_reply(z_owned_query_t *query,
                              zc_cgo_string_data_t keyexpr_data,
                              zc_cgo_bytes_data_t *payload_data,
                              zc_cgo_query_reply_options_t *opts) {
  z_owned_bytes_t payload, attachment;
  z_owned_encoding_t encoding;
  z_view_keyexpr_t keyexpr;
  z_view_keyexpr_from_substr_unchecked(&keyexpr, keyexpr_data.str_ptr,
                                       keyexpr_data.len);
  if (opts == NULL) {
    return z_query_reply(z_loan(*query), z_loan(keyexpr),
                         _create_moved_bytes_from_data(payload_data, &payload),
                         NULL);
  }
  z_query_reply_options_t options;
  z_query_reply_options_default(&options);
  if (opts->has_encoding) {
    options.encoding =
        _create_moved_encoding_from_data(&opts->encoding_data, &encoding);
  }
  if (opts->has_attachment) {
    options.attachment =
        _create_moved_bytes_from_data(&opts->attachment_data, &attachment);
  }
  options.is_express = opts->is_express;
  options.timestamp = opts->has_timestamp ? &opts->timestamp : NULL;
  options.source_info = opts->has_source_info ? &opts->source_info : NULL;
  return z_query_reply(z_loan(*query), z_loan(keyexpr),
                       _create_moved_bytes_from_data(payload_data, &payload),
                       &options);
}

z_result_t zc_cgo_query_reply_err(z_owned_query_t *query,
                                  zc_cgo_bytes_data_t *payload_data,
                                  zc_cgo_query_reply_err_options_t *opts) {
  z_owned_bytes_t payload;
  z_owned_encoding_t encoding;
  if (opts == NULL) {
    return z_query_reply_err(
        z_loan(*query), _create_moved_bytes_from_data(payload_data, &payload),
        NULL);
  }
  z_query_reply_err_options_t options;
  z_query_reply_err_options_default(&options);
  if (opts->has_encoding) {
    options.encoding =
        _create_moved_encoding_from_data(&opts->encoding_data, &encoding);
  }
  return z_query_reply_err(
      z_loan(*query), _create_moved_bytes_from_data(payload_data, &payload),
      &options);
}

z_result_t zc_cgo_query_reply_del(z_owned_query_t *query,
                                  zc_cgo_string_data_t keyexpr_data,
                                  zc_cgo_query_reply_del_options_t *opts) {
  z_view_keyexpr_t keyexpr;
  z_owned_bytes_t attachment;
  z_view_keyexpr_from_substr_unchecked(&keyexpr, keyexpr_data.str_ptr,
                                       keyexpr_data.len);
  if (opts == NULL) {
    return z_query_reply_del(z_loan(*query), z_loan(keyexpr), NULL);
  }
  z_query_reply_del_options_t options;
  z_query_reply_del_options_default(&options);
  if (opts->has_attachment) {
    options.attachment =
        _create_moved_bytes_from_data(&opts->attachment_data, &attachment);
  }
  options.is_express = opts->is_express;
  options.timestamp = opts->has_timestamp ? &opts->timestamp : NULL;
  options.source_info = opts->has_source_info ? &opts->source_info : NULL;
  return z_query_reply_del(z_loan(*query), z_loan(keyexpr), &options);
}

z_result_t zc_cgo_get(z_owned_session_t *session,
                      zc_cgo_string_data_t keyexpr_data, const char *params,
                      void *context, zc_cgo_get_options_t *opts) {
  z_view_keyexpr_t keyexpr;
  z_view_keyexpr_from_substr_unchecked(&keyexpr, keyexpr_data.str_ptr,
                                       keyexpr_data.len);
  z_owned_closure_reply_t closure;
  z_closure(&closure, zenohGetCallback, zenohGetDrop, context);
  if (opts == NULL) {
    return z_get(z_loan(*session), z_loan(keyexpr), params, z_move(closure),
                 NULL);
  }

  z_get_options_t options;
  z_get_options_default(&options);
  options.target = opts->target;
  options.consolidation = opts->consolidation;
  options.congestion_control = opts->congestion_control;
  options.is_express = opts->is_express;
  options.allowed_destination = opts->allowed_destination;
  options.accept_replies = opts->accept_replies;
  options.priority = opts->priority;
  options.timeout_ms = opts->timeout_ms;

  z_owned_bytes_t payload, attachment;
  z_owned_encoding_t encoding;
  z_owned_cancellation_token_t cancellation_token;

  if (opts->has_payload) {
    options.payload =
        _create_moved_bytes_from_data(&opts->payload_data, &payload);
  }
  if (opts->has_encoding) {
    options.encoding =
        _create_moved_encoding_from_data(&opts->encoding_data, &encoding);
  }
  if (opts->has_attachment) {
    options.attachment =
        _create_moved_bytes_from_data(&opts->attachment_data, &attachment);
  }
  if (opts->cancellation_token != NULL) {
    z_cancellation_token_clone(&cancellation_token,
                               z_loan(*opts->cancellation_token));
    options.cancellation_token = z_move(cancellation_token);
  }
  options.source_info = opts->has_source_info ? &opts->source_info : NULL;
  return z_get(z_loan(*session), z_loan(keyexpr), params, z_move(closure),
               &options);
}

z_result_t zc_cgo_liveliness_get(z_owned_session_t *session,
                                 zc_cgo_string_data_t keyexpr_data,
                                 void *context,
                                 zc_cgo_liveliness_get_options_t *opts) {
  z_view_keyexpr_t keyexpr;
  z_view_keyexpr_from_substr_unchecked(&keyexpr, keyexpr_data.str_ptr,
                                       keyexpr_data.len);
  z_owned_closure_reply_t closure;
  z_closure(&closure, zenohGetCallback, zenohGetDrop, context);

  if (opts == NULL) {
    return z_liveliness_get(z_loan(*session), z_loan(keyexpr), z_move(closure),
                            NULL);
  }

  z_liveliness_get_options_t options;
  z_liveliness_get_options_default(&options);

  options.timeout_ms = opts->timeout_ms;
  z_owned_cancellation_token_t cancellation_token;
  if (opts->cancellation_token != NULL) {
    z_cancellation_token_clone(&cancellation_token,
                               z_loan(*opts->cancellation_token));
    options.cancellation_token = z_move(cancellation_token);
  }
  return z_liveliness_get(z_loan(*session), z_loan(keyexpr), z_move(closure),
                          &options);
}

z_result_t zc_cgo_querier_get(z_owned_querier_t *querier, const char *params,
                              void *context,
                              zc_cgo_querier_get_options_t *opts) {
  z_owned_closure_reply_t closure;
  z_closure(&closure, zenohGetCallback, zenohGetDrop, context);

  if (opts == NULL) {
    return z_querier_get(z_loan(*querier), params, z_move(closure), NULL);
  }

  z_querier_get_options_t options;
  z_querier_get_options_default(&options);

  z_owned_bytes_t payload, attachment;
  z_owned_encoding_t encoding;
  z_owned_cancellation_token_t cancellation_token;
  if (opts->has_payload) {
    options.payload =
        _create_moved_bytes_from_data(&opts->payload_data, &payload);
  }
  if (opts->has_encoding) {
    options.encoding =
        _create_moved_encoding_from_data(&opts->encoding_data, &encoding);
  }
  if (opts->has_attachment) {
    options.attachment =
        _create_moved_bytes_from_data(&opts->attachment_data, &attachment);
  }
  if (opts->cancellation_token != NULL) {
    z_cancellation_token_clone(&cancellation_token,
                               z_loan(*opts->cancellation_token));
    options.cancellation_token = z_move(cancellation_token);
  }
  options.source_info = opts->has_source_info ? &opts->source_info : NULL;
  return z_querier_get(z_loan(*querier), params, z_move(closure), &options);
}

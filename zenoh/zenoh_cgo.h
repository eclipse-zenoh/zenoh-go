//
// Copyright (c) 2026 ZettaScale Technology
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

#ifndef ZENOH_CGO_H
#define ZENOH_CGO_H
#include "zenoh.h"

typedef struct {
  const char *str_ptr;
  size_t len;
} zc_cgo_string_data_t;

typedef struct {
  const void *data; // null if empty; z_loaned_bytes_t* if len == 0 && data !=
                    // NULL (non-contiguous); uint8_t* if len > 0
  size_t len; // 0 if data is z_loaned_bytes_t * or data is NULL, otherwise the
              // length of data uint8_t *
} zc_cgo_bytes_data_t;

typedef struct {
  zc_internal_encoding_data_t encoding;
  zc_cgo_string_data_t keyexpr;
  zc_cgo_bytes_data_t payload;
  zc_cgo_bytes_data_t attachment;
  const z_timestamp_t *timestamp;
  const z_source_info_t *source_info;
  z_sample_kind_t kind;
  z_reliability_t reliability;
} zc_cgo_sample_data_t;

typedef struct {
  zc_internal_encoding_data_t encoding;
  zc_cgo_string_data_t keyexpr;
  zc_cgo_bytes_data_t payload;
  zc_cgo_bytes_data_t attachment;
  zc_cgo_string_data_t params;
  z_owned_query_t query;
  const z_source_info_t *source_info;
  z_reply_keyexpr_t accepts_replies;
  bool has_encoding;
  bool has_payload;
  bool has_attachment;
} zc_cgo_query_data_t;

typedef struct {
  bool is_ok;
  zc_internal_encoding_data_t encoding;
  zc_cgo_string_data_t keyexpr;
  zc_cgo_bytes_data_t payload;
  zc_cgo_bytes_data_t attachment;
  const z_timestamp_t *timestamp;
  z_sample_kind_t kind;
  z_reliability_t reliability;
  const z_source_info_t *source_info;
} zc_cgo_reply_data_t;

typedef struct zc_cgo_get_options_t {
  zc_cgo_bytes_data_t payload_data;
  zc_internal_encoding_data_t encoding_data;
  zc_cgo_bytes_data_t attachment_data;
  uint64_t timeout_ms;
  z_owned_cancellation_token_t *cancellation_token;
  z_source_info_t source_info;
  z_congestion_control_t congestion_control;
  z_query_target_t target;
  z_query_consolidation_t consolidation;
  z_locality_t allowed_destination;
  z_reply_keyexpr_t accept_replies;
  z_priority_t priority;
  bool is_express;
  bool has_payload;
  bool has_encoding;
  bool has_attachment;
  bool has_source_info;
} zc_cgo_get_options_t;

typedef struct zc_cgo_querier_get_options_t {
  zc_cgo_bytes_data_t payload_data;
  zc_internal_encoding_data_t encoding_data;
  zc_cgo_bytes_data_t attachment_data;
  z_owned_cancellation_token_t *cancellation_token;
  z_source_info_t source_info;
  bool has_payload;
  bool has_encoding;
  bool has_attachment;
  bool has_source_info;
} zc_cgo_querier_get_options_t;

typedef struct zc_cgo_liveliness_get_options_t {
  uint64_t timeout_ms;
  z_owned_cancellation_token_t *cancellation_token;
} zc_cgo_liveliness_get_options_t;

zc_cgo_bytes_data_t zc_cgo_bytes_get_data(const z_loaned_bytes_t *bytes);
zc_cgo_string_data_t zc_cgo_string_get_data(const z_loaned_string_t *s);
zc_cgo_string_data_t zc_cgo_keyexpr_get_data(const z_loaned_keyexpr_t *keyexpr);

zc_cgo_sample_data_t zc_cgo_sample_get_data(z_loaned_sample_t *sample);
zc_cgo_query_data_t zc_cgo_query_get_data(z_loaned_query_t *query);
zc_cgo_reply_data_t zc_cgo_reply_get_data(z_loaned_reply_t *reply);

void zc_cgo_bytes_read_all(const z_loaned_bytes_t *bytes, uint8_t *out);

void zc_cgo_string_drop(z_owned_string_t *s);
void zc_cgo_encoding_drop(z_owned_encoding_t *e);
void zc_cgo_query_drop(z_owned_query_t *q);

extern void zenohSubscriberCallbackData(zc_cgo_sample_data_t sample,
                                        void *context);
extern void zenohSubscriberDrop(void *context);
void zenohSubscriberCallback(struct z_loaned_sample_t *sample, void *context);

extern void zenohQueryableCallbackData(zc_cgo_query_data_t query,
                                       void *context);
extern void zenohQueryableDrop(void *context);
void zenohQueryableCallback(struct z_loaned_query_t *query, void *context);

extern void zenohGetCallbackData(zc_cgo_reply_data_t query, void *context);
extern void zenohGetDrop(void *context);
void zenohGetCallback(struct z_loaned_reply_t *reply, void *context);

extern void zenohMatchingListenerDrop(void *context);
typedef const struct z_matching_status_t zc_cgo_const_matching_status;
extern void zenohMatchingListenerCallback(zc_cgo_const_matching_status *status,
                                          void *context);
extern void zenohMissListenerDrop(void *context);

extern void zenohTransportEventsCallback(z_loaned_transport_event_t *event,
                                         void *context);
extern void zenohTransportEventsDrop(void *context);
extern void zenohLinkEventsCallback(z_loaned_link_event_t *event,
                                    void *context);
extern void zenohLinkEventsDrop(void *context);

extern void zenohTransportCallback(z_loaned_transport_t *transport,
                                   void *context);
extern void zenohTransportDrop(void *context);
extern void zenohLinkCallback(z_loaned_link_t *link, void *context);
extern void zenohLinkDrop(void *context);

// Wrapper that returns false when Z_FEATURE_SHARED_MEMORY is not compiled in.
bool zc_cgo_transport_is_shm(const z_loaned_transport_t *transport);

// Wrapper that dispatches to zc_internal_create_transport_shm when shared memory is compiled in,
// falling back to zc_internal_create_transport otherwise.
void zc_cgo_create_transport(z_owned_transport_t *this_, z_id_t zid,
                              z_whatami_t whatami, bool is_qos,
                              bool is_multicast, bool is_shm);

typedef struct zc_cgo_put_options_t {
  zc_internal_encoding_data_t encoding_data;
  zc_cgo_bytes_data_t attachment_data;
  z_timestamp_t timestamp;
  z_locality_t allowed_destination;
  z_source_info_t source_info;
  z_congestion_control_t congestion_control;
  z_priority_t priority;
  z_reliability_t reliability;
  bool is_express;
  bool has_encoding;
  bool has_attachment;
  bool has_timestamp;
  bool has_source_info;
} zc_cgo_put_options_t;

typedef struct zc_cgo_delete_options_t {
  z_timestamp_t timestamp;
  z_locality_t allowed_destination;
  z_congestion_control_t congestion_control;
  z_priority_t priority;
  z_reliability_t reliability;
  bool is_express;
  bool has_timestamp;
} zc_cgo_delete_options_t;

typedef struct zc_cgo_publisher_put_options_t {
  zc_internal_encoding_data_t encoding_data;
  zc_cgo_bytes_data_t attachment_data;
  z_timestamp_t timestamp;
  z_source_info_t source_info;
  bool has_encoding;
  bool has_attachment;
  bool has_timestamp;
  bool has_source_info;
} zc_cgo_publisher_put_options_t;

typedef struct zc_cgo_publisher_delete_options_t {
  z_timestamp_t timestamp;
  bool has_timestamp;
} zc_cgo_publisher_delete_options_t;

typedef struct zc_cgo_query_reply_options_t {
  zc_internal_encoding_data_t encoding_data;
  zc_cgo_bytes_data_t attachment_data;
  z_timestamp_t timestamp;
  z_source_info_t source_info;
  bool is_express;
  bool has_encoding;
  bool has_attachment;
  bool has_timestamp;
  bool has_source_info;
} zc_cgo_query_reply_options_t;

typedef struct zc_cgo_query_reply_del_options_t {
  zc_cgo_bytes_data_t attachment_data;
  z_timestamp_t timestamp;
  z_source_info_t source_info;
  bool is_express;
  bool has_attachment;
  bool has_timestamp;
  bool has_source_info;
} zc_cgo_query_reply_del_options_t;

typedef struct zc_cgo_query_reply_err_options_t {
  zc_internal_encoding_data_t encoding_data;
  bool has_encoding;
} zc_cgo_query_reply_err_options_t;

z_result_t zc_cgo_publisher_put(z_owned_publisher_t *publisher,
                                zc_cgo_bytes_data_t *payload_data,
                                zc_cgo_publisher_put_options_t *opts);
z_result_t zc_cgo_publisher_delete(z_owned_publisher_t *publisher,
                                   zc_cgo_publisher_delete_options_t *opts);
z_result_t zc_cgo_put(z_owned_session_t *session,
                      zc_cgo_string_data_t keyexpr_data,
                      zc_cgo_bytes_data_t *payload_data,
                      zc_cgo_put_options_t *opts);
z_result_t zc_cgo_delete(z_owned_session_t *session,
                         zc_cgo_string_data_t keyexpr_data,
                         zc_cgo_delete_options_t *opts);
z_result_t zc_cgo_query_reply(z_owned_query_t *query,
                              zc_cgo_string_data_t keyexpr_data,
                              zc_cgo_bytes_data_t *payload_data,
                              zc_cgo_query_reply_options_t *opts);
z_result_t zc_cgo_query_reply_err(z_owned_query_t *query,
                                  zc_cgo_bytes_data_t *payload_data,
                                  zc_cgo_query_reply_err_options_t *opts);
z_result_t zc_cgo_query_reply_del(z_owned_query_t *query,
                                  zc_cgo_string_data_t keyexpr_data,
                                  zc_cgo_query_reply_del_options_t *opts);
z_result_t zc_cgo_get(z_owned_session_t *session,
                      zc_cgo_string_data_t keyexpr_data, const char *params,
                      void *context, zc_cgo_get_options_t *opts);
z_result_t zc_cgo_liveliness_get(z_owned_session_t *session,
                                 zc_cgo_string_data_t keyexpr_data,
                                 void *context,
                                 zc_cgo_liveliness_get_options_t *opts);
z_result_t zc_cgo_querier_get(z_owned_querier_t *querier, const char *params,
                              void *context,
                              zc_cgo_querier_get_options_t *opts);
#endif

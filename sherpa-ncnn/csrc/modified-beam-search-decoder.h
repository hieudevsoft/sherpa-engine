/**
 * Copyright (c)  2022  Xiaomi Corporation (authors: Fangjun Kuang)
 * Copyright (c)  2022                     (Pingfeng Luo)
 *
 * See LICENSE for clarification regarding multiple authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

#ifndef SHERPA_NCNN_CSRC_MODIFIED_BEAM_SEARCH_DECODER_H_
#define SHERPA_NCNN_CSRC_MODIFIED_BEAM_SEARCH_DECODER_H_

#include <vector>

#include "mat.h"  // NOLINT
#include "sherpa-ncnn/csrc/decoder.h"
#include "sherpa-ncnn/csrc/model.h"

namespace sherpa_ncnn {

class ModifiedBeamSearchDecoder : public Decoder {
 public:
  ModifiedBeamSearchDecoder(Model *model, int32_t num_active_paths)
      : model_(model), num_active_paths_(num_active_paths) {}

  DecoderResult GetEmptyResult() const override;

  void StripLeadingBlanks(DecoderResult *r) const override;

  void Decode(ncnn::Mat encoder_out, DecoderResult *result) override;

 private:
  ncnn::Mat BuildDecoderInput(const std::vector<Hypothesis> &hyps) const;

 private:
  Model *model_;  // not owned
  int32_t num_active_paths_;
};

}  // namespace sherpa_ncnn

#endif  // SHERPA_NCNN_CSRC_MODIFIED_BEAM_SEARCH_DECODER_H_

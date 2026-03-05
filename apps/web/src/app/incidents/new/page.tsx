import Layout from "@/layout";
import CreateEditIncident, {
  type IncidentFormValues,
} from "../components/create-edit-form";
import { BackButton } from "@/components/back-button";
import {
  getIncidentsQueryKey,
  postIncidentsMutation,
} from "@/api/@tanstack/react-query.gen";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { commonMutationErrorHandler } from "@/lib/utils";
import { useNavigate } from "react-router-dom";
import { useLocalizedTranslation } from "@/hooks/useTranslation";

const NewIncident = () => {
  const { t } = useLocalizedTranslation();
  const queryClient = useQueryClient();
  const navigate = useNavigate();

  const createIncidentMutation = useMutation({
    ...postIncidentsMutation(),
    onSuccess: () => {
      toast.success(t("incidents.created_success"));
      queryClient.invalidateQueries({ queryKey: getIncidentsQueryKey() });
      navigate("/incidents");
    },
    onError: commonMutationErrorHandler(t("common.error")),
  });

  const handleSubmit = (data: IncidentFormValues) => {
    createIncidentMutation.mutate({
      body: {
        status_page_id: data.statusPageId,
        title: data.title,
        content: data.content,
        style: data.style,
      },
    });
  };

  return (
    <Layout pageName={t("incidents.create")}>
      <BackButton to="/incidents" />
      <CreateEditIncident onSubmit={handleSubmit} />
    </Layout>
  );
};

export default NewIncident;

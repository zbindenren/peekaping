import { useParams, useNavigate } from "react-router-dom";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { BackButton } from "@/components/back-button";
import {
  getIncidentsByIdOptions,
  getIncidentsByIdQueryKey,
  getIncidentsQueryKey,
  patchIncidentsByIdMutation,
} from "@/api/@tanstack/react-query.gen";
import Layout from "@/layout";
import CreateEditIncident, {
  type IncidentFormValues,
} from "../components/create-edit-form";
import { toast } from "sonner";
import { commonMutationErrorHandler } from "@/lib/utils";
import { useLocalizedTranslation } from "@/hooks/useTranslation";

const EditIncident = () => {
  const { t } = useLocalizedTranslation();
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const queryClient = useQueryClient();

  const { data, isLoading, error } = useQuery({
    ...getIncidentsByIdOptions({
      path: { id: id! },
    }),
    enabled: !!id,
  });

  const incident = data?.data;

  const updateIncidentMutation = useMutation({
    ...patchIncidentsByIdMutation(),
    onSuccess: () => {
      toast.success(t("incidents.updated_success"));
      queryClient.removeQueries({ queryKey: getIncidentsQueryKey() });
      queryClient.removeQueries({
        queryKey: getIncidentsByIdQueryKey({
          path: { id: id! },
        }),
      });
      navigate("/incidents");
    },
    onError: commonMutationErrorHandler(t("common.error")),
  });

  const handleSubmit = (data: IncidentFormValues) => {
    updateIncidentMutation.mutate({
      path: { id: id! },
      body: {
        title: data.title,
        content: data.content,
        style: data.style,
      },
    });
  };

  if (isLoading)
    return (
      <Layout pageName={t("incidents.edit")}>{t("common.loading")}</Layout>
    );
  if (error || !data?.data)
    return (
      <Layout pageName={t("incidents.edit")}>{t("common.error")}</Layout>
    );

  const initialValues: IncidentFormValues = {
    statusPageId: incident?.status_page_id || "",
    title: incident?.title || "",
    content: incident?.content || "",
    style: (incident?.style as IncidentFormValues["style"]) || "warning",
  };

  return (
    <Layout pageName={t("incidents.edit")}>
      <BackButton to="/incidents" />
      <CreateEditIncident
        initialValues={initialValues}
        mode="edit"
        onSubmit={handleSubmit}
      />
    </Layout>
  );
};

export default EditIncident;

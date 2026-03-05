import { useMemo } from "react";
import { zodResolver } from "@hookform/resolvers/zod";
import {
  Form,
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form";
import { z } from "zod";
import { useForm } from "react-hook-form";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { useQuery } from "@tanstack/react-query";
import { getStatusPagesOptions } from "@/api/@tanstack/react-query.gen";
import { useLocalizedTranslation } from "@/hooks/useTranslation";

const incidentSchema = z.object({
  statusPageId: z.string().min(1, "incidents.form.status_page_required"),
  title: z.string().min(1, "incidents.form.title_required"),
  content: z.string().optional(),
  style: z.enum(["info", "warning", "danger"]),
});

export type IncidentFormValues = z.infer<typeof incidentSchema>;

const defaultValues: IncidentFormValues = {
  statusPageId: "",
  title: "",
  content: "",
  style: "warning",
};

export default function CreateEditIncident({
  initialValues = defaultValues,
  isLoading = false,
  mode = "create",
  onSubmit,
}: {
  initialValues?: IncidentFormValues;
  isLoading?: boolean;
  mode?: "create" | "edit";
  onSubmit: (data: IncidentFormValues) => void;
}) {
  const { t } = useLocalizedTranslation();

  const { data: statusPagesData } = useQuery({
    ...getStatusPagesOptions(),
  });

  const statusPages = statusPagesData?.data || [];

  const STYLE_OPTIONS = useMemo(
    () => [
      { value: "info", label: t("incidents.style.info") },
      { value: "warning", label: t("incidents.style.warning") },
      { value: "danger", label: t("incidents.style.danger") },
    ],
    [t]
  );

  const form = useForm<IncidentFormValues>({
    resolver: zodResolver(incidentSchema),
    defaultValues: initialValues,
  });

  const handleSubmit = (data: IncidentFormValues) => {
    onSubmit(data);
  };

  return (
    <div className="flex flex-col gap-6 max-w-[800px]">
      <CardTitle className="text-xl">
        {mode === "edit" ? t("incidents.edit") : t("incidents.create")}
      </CardTitle>

      <Form {...form}>
        <form onSubmit={form.handleSubmit(handleSubmit)} className="space-y-6">
          {mode === "create" && (
            <FormField
              control={form.control}
              name="statusPageId"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>{t("incidents.form.status_page_label")}</FormLabel>
                  <Select onValueChange={field.onChange} value={field.value}>
                    <FormControl>
                      <SelectTrigger>
                        <SelectValue
                          placeholder={t(
                            "incidents.form.select_status_page"
                          )}
                        />
                      </SelectTrigger>
                    </FormControl>
                    <SelectContent>
                      {statusPages.map((sp) => (
                        <SelectItem key={sp.id} value={sp.id || ""}>
                          {sp.title}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                  <FormMessage />
                </FormItem>
              )}
            />
          )}

          <FormField
            control={form.control}
            name="title"
            render={({ field }) => (
              <FormItem>
                <FormLabel>{t("incidents.form.title_label")}</FormLabel>
                <FormControl>
                  <Input
                    placeholder={t("incidents.form.title_placeholder")}
                    {...field}
                  />
                </FormControl>
                <FormMessage />
              </FormItem>
            )}
          />

          <FormField
            control={form.control}
            name="content"
            render={({ field }) => (
              <FormItem>
                <FormLabel>{t("incidents.form.content_label")}</FormLabel>
                <FormControl>
                  <Textarea
                    placeholder={t("incidents.form.content_placeholder")}
                    className="min-h-[100px]"
                    {...field}
                  />
                </FormControl>
                <FormDescription>
                  {t("incidents.form.markdown_supported")}
                </FormDescription>
                <FormMessage />
              </FormItem>
            )}
          />

          <FormField
            control={form.control}
            name="style"
            render={({ field }) => (
              <FormItem>
                <FormLabel>{t("incidents.form.style_label")}</FormLabel>
                <Select onValueChange={field.onChange} value={field.value}>
                  <FormControl>
                    <SelectTrigger>
                      <SelectValue />
                    </SelectTrigger>
                  </FormControl>
                  <SelectContent>
                    {STYLE_OPTIONS.map((option) => (
                      <SelectItem key={option.value} value={option.value}>
                        {option.label}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
                <FormMessage />
              </FormItem>
            )}
          />

          <div className="flex gap-2 pt-4">
            <Button type="submit" disabled={isLoading}>
              {isLoading ? t("common.saving") : t("common.save")}
            </Button>
          </div>
        </form>
      </Form>
    </div>
  );
}
